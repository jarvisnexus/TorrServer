package settings

import (
	"path/filepath"
	"strings"

	bolt "go.etcd.io/bbolt"
	"server/log"
)

type TDB struct {
	Path     string
	ReadOnly bool
	db       *bolt.DB
}

func NewTDB(path string, readOnly bool) *TDB {
	db, err := bolt.Open(filepath.Join(path, "config.db"), 0666, nil)
	if err != nil {
		log.TLogln(err)
		return nil
	}

	tdb := new(TDB)
	tdb.db = db
	tdb.Path = path
	tdb.ReadOnly = readOnly
	return tdb
}

func (v *TDB) CloseDB() {
	if v.db != nil {
		v.db.Close()
		v.db = nil
	}
}

func (v *TDB) Get(xpath, name string) []byte {
	spath := strings.Split(xpath, "/")
	if len(spath) == 0 {
		return nil
	}
	var ret []byte
	err := v.db.View(func(tx *bolt.Tx) error {
		buckt := tx.Bucket([]byte(spath[0]))
		if buckt == nil {
			return nil
		}

		for i, p := range spath {
			if i == 0 {
				continue
			}
			buckt = buckt.Bucket([]byte(p))
			if buckt == nil {
				return nil
			}
		}

		ret = buckt.Get([]byte(name))
		return nil
	})

	if err != nil {
		log.TLogln("Error get sets", xpath+"/"+name, ", error:", err)
	}

	return ret
}

func (v *TDB) Set(xpath, name string, value []byte) {
	if v.ReadOnly {
		return
	}

	spath := strings.Split(xpath, "/")
	if len(spath) == 0 {
		return
	}
	err := v.db.Update(func(tx *bolt.Tx) error {
		buckt, err := tx.CreateBucketIfNotExists([]byte(spath[0]))
		if err != nil {
			return err
		}

		for i, p := range spath {
			if i == 0 {
				continue
			}
			buckt, err = buckt.CreateBucketIfNotExists([]byte(p))
			if err != nil {
				return err
			}
		}

		return buckt.Put([]byte(name), value)
	})

	if err != nil {
		log.TLogln("Error put sets", xpath+"/"+name, ", error:", err)
		log.TLogln("value:", value)
	}

	return
}

func (v *TDB) List(xpath string) []string {
	spath := strings.Split(xpath, "/")
	if len(spath) == 0 {
		return nil
	}
	var ret []string
	err := v.db.View(func(tx *bolt.Tx) error {
		buckt := tx.Bucket([]byte(spath[0]))
		if buckt == nil {
			return nil
		}

		for i, p := range spath {
			if i == 0 {
				continue
			}
			buckt = buckt.Bucket([]byte(p))
			if buckt == nil {
				return nil
			}
		}

		buckt.ForEach(func(_, v []byte) error {
			if len(v) > 0 {
				ret = append(ret, string(v))
			}
			return nil
		})

		return nil
	})

	if err != nil {
		log.TLogln("Error list sets", xpath, ", error:", err)
	}

	return ret
}

func (v *TDB) Rem(xpath, name string) {
	if v.ReadOnly {
		return
	}

	spath := strings.Split(xpath, "/")
	if len(spath) == 0 {
		return
	}
	err := v.db.Update(func(tx *bolt.Tx) error {
		buckt := tx.Bucket([]byte(spath[0]))
		if buckt == nil {
			return nil
		}

		for i, p := range spath {
			if i == 0 {
				continue
			}
			buckt = buckt.Bucket([]byte(p))
			if buckt == nil {
				return nil
			}
		}

		return buckt.Delete([]byte(name))
	})

	if err != nil {
		log.TLogln("Error rem sets", xpath+"/"+name, ", error:", err)
	}

	return
}
