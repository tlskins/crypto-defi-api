package store

import (
	"github.com/globalsign/mgo"

	"github.com/my_projects/sol-arb-api/api"
)

// Store

type M api.M

type Store struct {
	m      *mgo.Session
	dbname string
}

func NewStore(mongoDBName, mongoHost, mongoUser, mongoPwd string) (*Store, error) {
	mongoClient, err := api.NewClientV2(mongoHost, mongoUser, mongoPwd)
	if err != nil {
		return nil, err
	}

	return &Store{
		m:      mongoClient,
		dbname: mongoDBName,
	}, nil
}

func (s *Store) DB() (*mgo.Session, *mgo.Database) {
	sess := s.m.Copy()
	return sess, sess.DB(s.dbname)
}

func (s *Store) C(colName string) (*mgo.Session, *mgo.Collection) {
	sess := s.m.Copy()
	return sess, sess.DB(s.dbname).C(colName)
}

func (s *Store) Close() {
	s.m.Close()
}
