package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

func GetAllGroups(context *Context) ([]Group, error) {
	log.Debug("Getting all groups")
	dbinfo := getConnectionString(context)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err)
	defer db.Close()

	rows, err := db.Query("SELECT groupId, oauthId, oauthSecret FROM groupinfo")
	checkErr(err)

	groups := []Group{}

	for rows.Next() {
		var groupId int
		var oauthId string
		var oauthSecret string
		err = rows.Scan(&groupId, &oauthId, &oauthSecret)
		checkErr(err)
		g := Group{groupId: groupId, oauthId: oauthId, oauthSecret: oauthSecret}

		groups = append(groups, g)
	}

	return groups, err
}

func GetGroup(context *Context, groupId int) (*Group, error) {
	log.Debugf("Getting group %d", groupId)
	dbinfo := getConnectionString(context)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err)
	defer db.Close()

	const query = `SELECT groupid, oauthId, oauthSecret, threshold from groupinfo where groupid = $1`
	var retval Group
	err = db.QueryRow(query, groupId).Scan(
		&retval.groupId, &retval.oauthId, &retval.oauthSecret, &retval.threshold)

	return &retval, err
}

func GetGroupByOauthId(context *Context, oauthId string) (*Group, error) {
	log.Debugf("Getting group by oauthId")
	dbinfo := getConnectionString(context)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err)
	defer db.Close()

	const query = `SELECT groupid, oauthId, oauthSecret, threshold from groupinfo where oauthId = $1`
	var retval Group
	err = db.QueryRow(query, oauthId).Scan(
		&retval.groupId, &retval.oauthId, &retval.oauthSecret, &retval.threshold)

	return &retval, err
}

func DeleteGroup(context *Context, oauthId string) (*Group, error) {
	dbinfo := getConnectionString(context)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err)
	defer db.Close()

	var retval Group
	const query = `DELETE FROM groupinfo where oauthId = $1 RETURNING groupid, oauthId, oauthSecret`
	err = db.QueryRow(query, oauthId).Scan(
		&retval.groupId, &retval.oauthId, &retval.oauthSecret)
	return &retval, err
}

func AddGroup(context *Context, groupId int, oauthId string, oauthSecret string, threshold int) error {
	log.Debugf("Adding group %d", groupId)
	dbinfo := getConnectionString(context)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err)
	defer db.Close()

	var returned_gid int
	const query = `INSERT INTO groupinfo(groupId,oauthId,oauthSecret,threshold) VALUES($1,$2,$3,$4) ON CONFLICT(groupid) DO UPDATE SET oauthId=$2,oauthSecret=$3,threshold=$4 RETURNING groupId`
	err = db.QueryRow(query, groupId, oauthId, oauthSecret, threshold).Scan(&returned_gid)
	return err
}

func UpdateThreshold(context *Context, oauthId string, threshold int) (*Group, error) {
	log.Debugf("UpdateThreshold of oauthid-%s to %d", oauthId, threshold)
	dbinfo := getConnectionString(context)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err)
	defer db.Close()

	var retval Group
	const query = `UPDATE groupinfo SET threshold=$1 where oauthId=$2  RETURNING groupid, oauthId, oauthSecret, threshold`
	err = db.QueryRow(query, threshold, oauthId).Scan(
		&retval.groupId, &retval.oauthId, &retval.oauthSecret, &retval.threshold)
	return &retval, err
}

func getConnectionString(context *Context) string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", context.pguser, context.pgpass, context.pghost, context.pgdatabase)
}
