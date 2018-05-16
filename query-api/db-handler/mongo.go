/*
 * Copyright (c) 2018 Juniper Networks, Inc. All rights reserved.
 *
 * file:    mongo.go
 * details: Deals with the setting up connection with Mongo DB
 *
 */
package dbhandler

import (
	"net/http"

	mgo "gopkg.in/mgo.v2"

	opts "github.com/Juniper/collector/query-api/options"
)

type MongoDBHandler struct {
	mgoSession *mgo.Session
	mgoDB      *mgo.Database
	mgoDBs     []*mgo.Database
}

func (mg *MongoDBHandler) connectToMongo() *mgo.Session {
	var (
		mongo = opts.MongoIP + ":" + opts.MongoPort
	)
	opts.Logger.Println("Connecting Mongo", mongo)
	dialInfo := &mgo.DialInfo{
		Addrs:    []string{mongo},
		Username: opts.MongoUserName,
		Password: opts.MongoUserPasswd,
	}
	mgoSession, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		mg.connectToMongo()
	}
	return mgoSession
}

/*
func (mg *MongoDBHandler) ensureIndex() error {
}
*/
func (mg *MongoDBHandler) setup(mux *http.ServeMux) error {
	if opts.UseDatabase != opts.UseDatabaseMongo {
		opts.Logger.Println("Database Mongo selection is disabled")
		return nil
	}
	mg.mgoSession = mg.connectToMongo()
	thisDB := mg.mgoSession.Copy()
	mg.mgoDB = thisDB.DB(opts.DBFlows)
	if opts.DoNeedQuerySplit {
		mg.mgoDBs = make([]*mgo.Database, opts.MongoConnectionPoolLen)
		for i := 0; i < opts.MongoConnectionPoolLen; i++ {
			thisDB := mg.mgoSession.Copy()
			mg.mgoDBs[i] = thisDB.DB(opts.DBFlows)
		}
	}
	mg.RegisterHandlers(mux)
	return nil
}
