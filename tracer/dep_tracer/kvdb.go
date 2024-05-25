package dep_tracer

import (
    "os"
    "github.com/syndtr/goleveldb/leveldb"
    riak "github.com/basho/riak-go-client"
)

type DB interface {
    Get(key []byte, optional bool) []byte
    Set(key, value []byte)
    Delete(key []byte)
    DumpAllDebug() map[string][]byte
}

func NewDB(engine, root, name string) DB {
    switch engine {
    case "leveldb":
        return NewLevelDB(root, name)
    case "riak":
        return NewRiakDB(root, name)
    default:
        panic("unknown engine")
    }
}


type LevelDB struct {
    db *leveldb.DB
}

func NewLevelDB(root, name string) LevelDB {
    err := os.MkdirAll(root, os.ModePerm)
    if err != nil {
        panic(nil)
    }
    path := root + "/" + name
    db := LevelDB{}
    db.db, err = leveldb.OpenFile(path, nil)
    if err != nil {
        panic(err)
    }
    return db
}

func (db LevelDB) Get(key []byte, optional bool) []byte {
    val, err := db.db.Get(key, nil)
    if err != nil && err != leveldb.ErrNotFound {
        panic(err)
    }
    if val != nil {
        return val
    } else {
        if optional {
            return nil
        } else {
            panic("key not found")
        }
    }
}

func (db LevelDB) Set(key, value []byte) {
    err := db.db.Put(key, value, nil)
    if err != nil {
        panic(err)
    }
}

func (db LevelDB) Delete(key []byte) {
    err := db.db.Delete(key, nil)
    if err != nil {
        panic(err)
    }
}

func (db LevelDB) DumpAllDebug() map[string][]byte {
    res := map[string][]byte{}
    iter := db.db.NewIterator(nil, nil)
    for iter.Next() {
        key := string(iter.Key())
        value := iter.Value()
        res[key] = value
    }
    iter.Release()
    err := iter.Error()
    if err != nil {
        panic(err)
    }
    return res
}



var riakDB *riak.Client = nil
type RiakDB struct {
    name string
}

func NewRiakDB(root, name string) RiakDB {
    if riakDB == nil {
        var err error
        riakDB, err = riak.NewClient(&riak.NewClientOptions{
            RemoteAddresses: []string{root},
        })
        if err != nil {
            panic(err)
        }
    }
    db := RiakDB{
        name: name,
    }
    return db
}

func (db RiakDB) Get(key []byte, optional bool) []byte {
    cmd, err := riak.NewFetchValueCommandBuilder().
        WithBucket(db.name).
        WithKey(string(key)).
        Build()
    if err != nil {
        panic(err)
    }
  
    err = riakDB.Execute(cmd)
    if err != nil {
        panic(err)
    }

    fcmd := cmd.(*riak.FetchValueCommand)
    values := fcmd.Response.Values
    if len(values) < 1 {
        if optional {
            return nil
        } else {
            panic("key not found")
        }
    } else {
        return values[0].Value
    }
}

func (db RiakDB) Set(key, value []byte) {
    content := &riak.Object{
        Bucket:      db.name,
        Key:         string(key),
        ContentType: "application/octet-stream",
        Value:       value,
    }

    cmd, err := riak.NewStoreValueCommandBuilder().
        WithContent(content).
        Build()
    if err != nil {
        panic(err)
    }

    err = riakDB.Execute(cmd)
    if err != nil {
        panic(err)
    }
}

func (db RiakDB) Delete(key []byte) {
    cmd, err := riak.NewDeleteValueCommandBuilder().
        WithBucket(db.name).
        WithKey(string(key)).
        Build()
    if err != nil {
        panic(err)
    }

    err = riakDB.Execute(cmd)
    if err != nil {
        panic(err)
    }
}

func (db RiakDB) DumpAllDebug() map[string][]byte {
    panic("DumpAllDebug() not implemented")
}
