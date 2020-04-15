package main

import (
    "bytes"
    "encoding/gob"
    "encoding/json"
    "io/ioutil"
    "testing"

    "google.golang.org/protobuf/proto"
)

type AgentData struct {
    Hostname    string   `json:"hostname"`
    Status      string   `json:"status"`
    Timestamp   int      `json:"timestamp"`
    Lsns        []string `json:"lsns"`
}

func generateObject() *AgentData {
    return &AgentData{
        Hostname:   "10.64.6.138",
        Status:     "In Progress",
        Timestamp:  1282368345,
        Lsns:       []string{"16/B374D848", "16/B374D010"},
    }
}

func generateProtoBufObject() *BookProto {
    obj := generateObject()

    return &BookProto{
        Hostname:   *proto.String(obj.Hostname),
        Status:     *proto.String(obj.Status),
        Timestamp:  *proto.Int64(int64(obj.Timestamp)),
        Lsns:       obj.Lsns,
    }
}

func BenchmarkJSONMarshal(b *testing.B) {
    obj := generateObject()

    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        _, err := json.Marshal(obj)
        if err != nil {
            panic(err)
        }
    }
}

func BenchmarkJSONUnmarshal(b *testing.B) {
    out, err := json.Marshal(generateObject())
    if err != nil {
        panic(err)
    }

    obj := &AgentData{}

    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        err = json.Unmarshal(out, obj)
        if err != nil {
            panic(err)
        }
    }
}

func BenchmarkProtoBufMarshal(b *testing.B) {
    obj := generateProtoBufObject()

    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        _, err := proto.Marshal(obj)
        if err != nil {
            panic(err)
        }
    }
}

func BenchmarkProtoBufUnmarshal(b *testing.B) {
    out, err := proto.Marshal(generateProtoBufObject())
    if err != nil {
        panic(err)
    }

    obj := &BookProto{}

    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        err = proto.Unmarshal(out, obj)
        if err != nil {
            panic(err)
        }
    }
}

func BenchmarkGobMarshal(b *testing.B) {
    obj := generateObject()

    enc := gob.NewEncoder(ioutil.Discard)

    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        err := enc.Encode(obj)
        if err != nil {
            panic(err)
        }
    }
}

func BenchmarkGobUnmarshal(b *testing.B) {
    obj := generateObject()

    var buf bytes.Buffer
    enc := gob.NewEncoder(&buf)
    err := enc.Encode(obj)
    if err != nil {
        panic(err)
    }

    for n := 0; n < b.N; n++ {
        err = enc.Encode(obj)
        if err != nil {
            panic(err)
        }
    }

    dec := gob.NewDecoder(&buf)

    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        err = dec.Decode(&AgentData{})
        if err != nil {
            panic(err)
        }
    }
}
