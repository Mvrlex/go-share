package storage

import (
	"fmt"
	"io"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestStoringAndRetrieving(t *testing.T) {
	storage, err := NewPhysicalStorage("./shared1")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = storage.Close(); err != nil {
			t.Log(err)
		}
	}()
	key, err := storage.Store("my-test-file.txt", strings.NewReader("my contents"), "", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("value successfully stored as file, given key is:", key)
	value, err := storage.Get(key, "")
	if value.Name != "my-test-file.txt" {
		t.Fatalf("stored name did not match given name: %q != \"my-test-file.txt\"", value.Name)
	}
	out := make([]byte, 11)
	_, err = io.ReadFull(value, out)
	if err != nil {
		t.Fatal(err)
	}
	err = value.Close()
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "my contents" {
		t.Fatalf("returned value did not match input value: %q != \"my contents\"", string(out))
	}
	fmt.Println("value behind stored key is:", string(out))
}

func TestStoringAndRetrievingWithCustomPassword(t *testing.T) {
	storage, err := NewPhysicalStorage("./shared2")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = storage.Close(); err != nil {
			t.Log(err)
		}
	}()
	key, err := storage.Store("my-test-file.txt", strings.NewReader("my contents"), "password", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("value successfully stored as encrypted file, given key is:", key)
	value, err := storage.Get(key, "password")
	if value.Name != "my-test-file.txt" {
		t.Fatalf("stored name did not match given name: %q != \"my-test-file.txt\"", value.Name)
	}
	out := make([]byte, 11)
	_, err = io.ReadFull(value, out)
	if err != nil {
		t.Fatal(err)
	}
	err = value.Close()
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "my contents" {
		t.Fatalf("returned value did not match input value: %q != \"my contents\"", string(out))
	}
	fmt.Println("value behind stored key is:", string(out))
}

func TestStoringFileWithPasswordReceivingWithoutPassword(t *testing.T) {
	storage, err := NewPhysicalStorage("./shared3")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = storage.Close(); err != nil {
			t.Log(err)
		}
	}()
	key, err := storage.Store("my-test-file.txt", strings.NewReader("my contents"), "password", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("value successfully stored as encrypted file, given key is:", key)
	value, err := storage.Get(key, "")
	if err == nil {
		value.Close()
		t.Fatal("receiving encrypted file unencrypted should not be possible")
	}
}

func TestStoringFileWithoutPasswordReceivingWithPassword(t *testing.T) {
	storage, err := NewPhysicalStorage("./shared4")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = storage.Close(); err != nil {
			t.Log(err)
		}
	}()
	key, err := storage.Store("my-test-file.txt", strings.NewReader("my contents"), "", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("value successfully stored as encrypted file, given key is:", key)
	value, err := storage.Get(key, "lol")
	if err == nil {
		value.Close()
		t.Fatal("could get value even though the provided password is incorrect")
	}
}

func TestRemovingFile(t *testing.T) {
	storage, err := NewPhysicalStorage("./shared5")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = storage.Close(); err != nil {
			t.Log(err)
		}
	}()
	key, err := storage.Store("my-test-file.txt", strings.NewReader("my contents"), "", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	storage.Remove(key)
	value, err := storage.Get(key, "")
	if err == nil {
		value.Close()
		t.Fatal("receiving already deleted file should not be possible")
	}
}

func TestCreatingStorageWithDirtyDirectory(t *testing.T) {
	storage, err := NewPhysicalStorage("./shared6")
	if err != nil {
		t.Fatal(err)
	}
	storage.Store("my-test-file.txt", strings.NewReader("my contents"), "", time.Minute)
	storage, err = NewPhysicalStorage("./shared6")
	if err != nil {
		t.Fatal(err)
	}
	err = storage.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestFileSelfDestruct(t *testing.T) {
	storage, err := NewPhysicalStorage("./shared7")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = storage.Close(); err != nil {
			t.Log(err)
		}
	}()
	key, err := storage.Store("my-test-file.txt", strings.NewReader("my contents"), "", time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond * 100)
	runtime.GC()
	value, err := storage.Get(key, "")
	if err == nil {
		value.Close()
		t.Fatal("receiving self destructed file should not be possible")
	}
}

func TestGetFileInfo(t *testing.T) {
	storage, err := NewPhysicalStorage("./shared8")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = storage.Close(); err != nil {
			t.Log(err)
		}
	}()
	key, err := storage.Store("my-test-file.txt", strings.NewReader("my contents"), "", time.Minute)
	storedValueInfo := storage.Info(key)
	if storedValueInfo == nil {
		t.Fatal("value was successfully stored but no information is available")
	}
	if storedValueInfo.Bytes != 11 || storedValueInfo.RequiresPassword != false || storedValueInfo.Name != "my-test-file.txt" {
		t.Fatal("stored information did not match input")
	}
}

func TestGetSize(t *testing.T) {
	storage, err := NewPhysicalStorage("./shared9")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = storage.Close(); err != nil {
			t.Log(err)
		}
	}()
	_, err = storage.Store("my-test-file.txt", strings.NewReader("my contents"), "", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	_, err = storage.Store("my-test-file.txt", strings.NewReader("my contents"), "", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if storage.Size() != 22 {
		t.Fatal("returned size did not match size in storage")
	}
}

func TestFileGarbageCollection(t *testing.T) {
	storage, err := NewPhysicalStorage("./shared10")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = storage.Close(); err != nil {
			t.Log(err)
		}
	}()
	key, err := storage.Store("my-test-file.txt", strings.NewReader("my contents"), "", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	storedValue, err := storage.Get(key, "")
	defer storedValue.Close()
	storage.Remove(key)
	time.Sleep(time.Millisecond * 100)
	runtime.GC()
	_, err = io.ReadAll(storedValue)
	if err != nil {
		t.Fatal(err)
	}
}
