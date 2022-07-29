package store

import (
	"github.com/my_projects/sol-arb-api/api"
	t "github.com/my_projects/sol-arb-api/types"
	uuid "github.com/satori/go.uuid"
)

var colTokenTrackers = "tokenTrackers"

func (s *Store) GetTokenTrackers() (out []*t.TokenTracker, err error) {
	sess, c := s.C(colTokenTrackers)
	defer sess.Close()

	out = []*t.TokenTracker{}
	err = api.Find(c, &out, api.M{})
	return
}

func (s *Store) GetTokenTracker(id string) (out *t.TokenTracker, err error) {
	sess, c := s.C(colTokenTrackers)
	defer sess.Close()

	out = &t.TokenTracker{}
	err = api.FindOne(c, out, api.M{"_id": id})
	return
}

func (s *Store) GetTokenTrackersByDiscordId(discId string) (out []*t.TokenTracker, err error) {
	sess, c := s.C(colTokenTrackers)
	defer sess.Close()

	out = []*t.TokenTracker{}
	err = api.Find(c, &out, api.M{"discordId": discId})
	return
}

func (s *Store) UpsertTokenTracker(upsert *t.TokenTracker) (out *t.TokenTracker, err error) {
	sess, c := s.C(colTokenTrackers)
	defer sess.Close()

	if upsert.Id == "" {
		upsert.Id = uuid.NewV4().String()
	}

	out = &t.TokenTracker{}
	err = api.Upsert(c, out, M{"_id": upsert.Id}, M{"$set": upsert})
	return
}

func (s *Store) DeleteTokenTracker(id string) (err error) {
	sess, c := s.C(colTokenTrackers)
	defer sess.Close()

	return api.Remove(c, api.M{"_id": id})
}
