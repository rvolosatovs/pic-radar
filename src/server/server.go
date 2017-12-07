package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	bolt "github.com/coreos/bbolt"
	influx "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
)

type store struct {
	boltdb   *bolt.DB
	influxdb influx.Client
}

type Store struct {
	store
	Users   UserStore
	Queries QueryStore
}

type BoltConfig struct {
	Filename string
	Options  *bolt.Options
}

func NewStore(bc BoltConfig, ic influx.HTTPConfig) (*Store, error) {
	boltdb, err := bolt.Open(bc.Filename, 0666, bc.Options)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open BoltDB")
	}
	if err := boltdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("users"))
		return err
	}); err != nil {
		return nil, err
	}

	influxdb, err := influx.NewHTTPClient(ic)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open InfluxDB")
	}
	resp, err := influxdb.Query(influx.NewQuery("CREATE DATABASE requests", "", ""))
	if err != nil {
		return nil, err
	}
	if err = resp.Error(); err != nil {
		return nil, err
	}

	s := &Store{
		store: store{
			boltdb:   boltdb,
			influxdb: influxdb,
		},
	}
	s.Users = UserStore(s.store)
	s.Queries = QueryStore(s.store)
	return s, nil
}

type Image struct {
	Link string `json:"link"`
}

func ReadImage(r io.Reader) (*Image, error) {
	img := &Image{}
	if err := json.NewDecoder(r).Decode(img); err != nil {
		return nil, err
	}
	return img, nil
}

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Query    string `json:"query"`
}

func ReadUser(r io.Reader) (*User, error) {
	txt, er := ioutil.ReadAll(r)
	if er != nil {
		panic(er)
	}
	fmt.Println(string(txt))
	r = strings.NewReader(string(txt))

	u := &User{}
	if err := json.NewDecoder(r).Decode(u); err != nil {
		return nil, err
	}
	return u, nil
}

type UserStore store

func (s UserStore) Add(u *User) error {
	b, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return s.boltdb.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("users")).Put([]byte(u.Login), b)
	})
}

func (s UserStore) Exists(login string) (ok bool, err error) {
	return ok, s.boltdb.View(func(tx *bolt.Tx) error {
		ok = tx.Bucket([]byte("users")).Get([]byte(login)) != nil
		return nil
	})
}

func (s UserStore) Get(login string) (u *User, err error) {
	u = &User{}
	return u, s.boltdb.View(func(tx *bolt.Tx) error {
		return json.Unmarshal(tx.Bucket([]byte("users")).Get([]byte(login)), u)
	})
}

type Query struct {
	Timestamp time.Time     `json:"timestamp"`
	Duration  time.Duration `json:"duration"`
	Endpoint  string        `json:"endpoint"`
	RawQuery  string        `json:"query"`
	User      *User         `json:"user"`
	Address   string        `json:"address"`
}

type QueryStore store

func (s QueryStore) Add(q Query) error {
	bp, err := influx.NewBatchPoints(influx.BatchPointsConfig{
		Database: "requests",
	})
	tags := map[string]string{
		"address": q.Address,
	}
	if q.User != nil {
		tags["user"] = q.User.Login
	}

	p, err := influx.NewPoint(q.Endpoint,
		tags,
		map[string]interface{}{
			"query":    q.RawQuery,
			"duration": q.Duration,
		},
		q.Timestamp,
	)
	if err != nil {
		return err
	}
	bp.AddPoint(p)
	return s.influxdb.Write(bp)
}
