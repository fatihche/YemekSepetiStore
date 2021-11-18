package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)

type ModalSet struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}
type ModalGet struct {
	Key string `json:"Key"`
}
type MemoryStore struct {
	storeMap      map[string]string
	storeMapMutex sync.RWMutex
	storeChannel  chan ModalSet
	fileName      string
	perMinute     time.Duration
}

func newMemoryStore() (memoryStore *MemoryStore) {
	memoryStore = &MemoryStore{
		storeMap:      make(map[string]string),
		storeMapMutex: sync.RWMutex{},
		storeChannel:  make(chan ModalSet),
		fileName:      "store-data.json",
		perMinute:     2,
	}
	return
}

func (memoryStore *MemoryStore) set(w http.ResponseWriter, req *http.Request) {
	reqBody, _ := ioutil.ReadAll(req.Body)
	var modalSet ModalSet
	json.Unmarshal(reqBody, &modalSet)
	if modalSet.Key == "" || modalSet.Value == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	memoryStore.storeChannel <- modalSet
}

func (memoryStore *MemoryStore) get(w http.ResponseWriter, req *http.Request) {
	reqBody, _ := ioutil.ReadAll(req.Body)
	var modalGet ModalGet
	json.Unmarshal(reqBody, &modalGet)
	if modalGet.Key == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "%+v", string(memoryStore.getMemoryStore(modalGet)))
}

func (memoryStore *MemoryStore) setMemoryStore() {
	for {
		modalSet := <-memoryStore.storeChannel
		memoryStore.storeMap[modalSet.Key] = modalSet.Value
		fmt.Println(memoryStore.storeMap)
	}
}

func (memoryStore *MemoryStore) getMemoryStore(modalGet ModalGet) (value string) {
	memoryStore.storeMapMutex.RLock()
	value = memoryStore.storeMap[modalGet.Key]
	memoryStore.storeMapMutex.RUnlock()
	return
}

func (memoryStore *MemoryStore) readFile() {
	if !fileExists(memoryStore.fileName) {
		err := ioutil.WriteFile(memoryStore.fileName, []byte("{}"), 0755)
		if err != nil {
			fmt.Printf("Unable to write file: %v", err)
		}
	}
	data, err := ioutil.ReadFile(memoryStore.fileName)
	if err != nil {
		fmt.Println(err)
	}
	if err := json.Unmarshal(data, &memoryStore.storeMap); err != nil {
		panic(err)
	}
	_, err = json.Marshal(memoryStore.storeMap)
	if err != nil {
		panic(err)
	}
}
func (memoryStore *MemoryStore) writeFile() {
	for true {
		time.Sleep(memoryStore.perMinute * time.Minute)
		fmt.Println("asd")
		jsonString, _ := json.Marshal(memoryStore.storeMap)
		err := ioutil.WriteFile(memoryStore.fileName, jsonString, 0777)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func startMemoryStore() {
	fmt.Println("start")
	memoryStore := newMemoryStore()
	memoryStore.readFile()
	go memoryStore.writeFile()
	go memoryStore.setMemoryStore()
	http.HandleFunc("/set", memoryStore.set)
	http.HandleFunc("/get", memoryStore.get)
	http.ListenAndServe(":8081", nil)
}

func main() {
	startMemoryStore()
}
