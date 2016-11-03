package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

type MockDB struct {
	storage map[string]map[string]interface{}
}

func NewMockDB() *MockDB {
	return &MockDB{
		storage: make(map[string]map[string]interface{}),
	}
}

func (m *MockDB) hasSHA1(sha1 string) (bool, error) {
	var ok bool
	_, ok = m.storage[sha1]
	return ok, nil

}

func (m *MockDB) getTreeURLs(sha1 string) ([]string, error) {
	return []string{m.storage[sha1]["tree_urls"].(string)}, nil
}

func (m *MockDB) deleteTreeURLs(sha1 string) error {
	delete(m.storage, sha1)
	return nil
}

func (m *MockDB) insertTreeURLs(sha1, treeURLs string) error {
	if _, ok := m.storage[sha1]["tree_urls"]; !ok {
		m.storage[sha1] = make(map[string]interface{})
	}
	m.storage[sha1]["tree_urls"] = treeURLs
	return nil
}

func (m *MockDB) updateTreeURLs(sha1, treeURLs string) error {
	return m.insertTreeURLs(sha1, treeURLs)
}

func TestBadRequest(t *testing.T) {
	var (
		expectedMsg    = "test message\n"
		expectedStatus = http.StatusBadRequest
	)

	recorder := httptest.NewRecorder()
	badRequest(recorder, "test message")
	actualMsg := recorder.Body.String()
	actualStatus := recorder.Code

	if actualStatus != expectedStatus {
		t.Errorf("Status code was %d and should have been %d", actualStatus, expectedStatus)
	}

	if actualMsg != expectedMsg {
		t.Errorf("Expected message was '%s', but should have been '%s'", actualMsg, expectedMsg)
	}
}

func TestErrored(t *testing.T) {
	var (
		expectedMsg    = "test message\n"
		expectedStatus = http.StatusInternalServerError
	)

	recorder := httptest.NewRecorder()
	errored(recorder, "test message")
	actualMsg := recorder.Body.String()
	actualStatus := recorder.Code

	if actualStatus != expectedStatus {
		t.Errorf("Status code was %d and should have been %d", actualStatus, expectedStatus)
	}

	if actualMsg != expectedMsg {
		t.Errorf("Expected message was '%s', but should have been '%s'", actualMsg, expectedMsg)
	}
}

func TestNotFound(t *testing.T) {
	var (
		expectedMsg    = "test message\n"
		expectedStatus = http.StatusNotFound
	)

	recorder := httptest.NewRecorder()
	notFound(recorder, "test message")
	actualMsg := recorder.Body.String()
	actualStatus := recorder.Code

	if actualStatus != expectedStatus {
		t.Errorf("Status code was %d and should have been %d", actualStatus, expectedStatus)
	}

	if actualMsg != expectedMsg {
		t.Errorf("Expected message was '%s', but should have been '%s'", actualMsg, expectedMsg)
	}
}

func TestValidSHA1(t *testing.T) {
	var (
		err      error
		valid    bool
		goodSHA1 = "60e3da2efd886074e28e44d48cc642f84c25b140"
		badSHA1  = "60e3da2efd886074e28e44d48cc64"
	)

	if valid, err = validSHA1(goodSHA1); err != nil {
		t.Error(err)
	}
	if !valid {
		t.Errorf("SHA1 '%s' was reported as invalid when it is valid", goodSHA1)
	}

	if valid, err = validSHA1(badSHA1); err != nil {
		t.Error(err)
	}
	if !valid {
		t.Errorf("SHA1 '%s' was reported as invalid", badSHA1)
	}
}

func TestNew(t *testing.T) {
	mock := NewMockDB()
	n := New(mock)
	if n == nil {
		t.Error("New returned nil instead of a *TreeURLs")
	}
}

func TestGreeting(t *testing.T) {
	mock := NewMockDB()
	n := New(mock)

	server := httptest.NewServer(n.router)
	defer server.Close()

	res, err := http.Get(server.URL)
	if err != nil {
		t.Error(err)
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Error(err)
	}

	actualBody := string(bodyBytes)
	expectedBody := "Hello from tree-urls."

	expectedStatus := http.StatusOK
	actualStatus := res.StatusCode

	if actualBody != expectedBody {
		t.Errorf("Body of the response was '%s' instead of '%s'", actualBody, expectedBody)
	}

	if actualStatus != expectedStatus {
		t.Errorf("Status of the response was %d instead of %d", actualStatus, expectedStatus)
	}
}

func TestGet(t *testing.T) {
	sha1 := "60e3da2efd886074e28e44d48cc642f84c25b140"
	expectedBody := `[{"label":"tree_0","url":"http://portnoy.iplantcollaborative.org/view/tree/3727f35cc7125567492cab69850f6473"}]`

	mock := NewMockDB()
	if err := mock.insertTreeURLs(sha1, expectedBody); err != nil {
		t.Error(err)
	}

	n := New(mock)
	server := httptest.NewServer(n.router)
	defer server.Close()

	sha1URL := fmt.Sprintf("%s/%s", server.URL, sha1)
	res, err := http.Get(sha1URL)
	if err != nil {
		t.Error(err)
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Error(err)
	}

	actualBody := string(bodyBytes)

	expectedStatus := http.StatusOK
	actualStatus := res.StatusCode

	if actualBody != expectedBody {
		t.Errorf("Body of the response was '%s' instead of '%s'", actualBody, expectedBody)
	}

	if actualStatus != expectedStatus {
		t.Errorf("Status of the response was %d instead of %d", actualStatus, expectedStatus)
	}
}

func TestPutInsert(t *testing.T) {
	sha1 := "60e3da2efd886074e28e44d48cc642f84c25b140"
	treeURL := `[{"label":"tree_0","url":"http://portnoy.iplantcollaborative.org/view/tree/3727f35cc7125567492cab69850f6473"}]`

	mock := NewMockDB()
	n := New(mock)
	server := httptest.NewServer(n.router)
	defer server.Close()

	sha1URL := fmt.Sprintf("%s/%s", server.URL, sha1)
	httpClient := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, sha1URL, strings.NewReader(treeURL))
	if err != nil {
		t.Error(err)
	}
	res, err := httpClient.Do(req)
	if err != nil {
		t.Error(err)
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Error(err)
	}

	var parsed map[string]string
	err = json.Unmarshal(bodyBytes, &parsed)
	if err != nil {
		t.Error(err)
	}

	if _, ok := parsed["tree_urls"]; !ok {
		t.Error("Parsed response did not have a top-level 'tree_urls' key")
	}

	if parsed["tree_urls"] != treeURL {
		t.Errorf("Put returned '%s' as the tree URL instead of '%s'", parsed["tree_urls"], treeURL)
	}
}

func TestPutUpdate(t *testing.T) {
	sha1 := "60e3da2efd886074e28e44d48cc642f84c25b140"
	treeURL := `[{"label":"tree_0","url":"http://portnoy.iplantcollaborative.org/view/tree/3727f35cc7125567492cab69850f6473"}]`

	mock := NewMockDB()
	if err := mock.insertTreeURLs(sha1, treeURL); err != nil {
		t.Error(err)
	}

	n := New(mock)
	server := httptest.NewServer(n.router)
	defer server.Close()

	sha1URL := fmt.Sprintf("%s/%s", server.URL, sha1)
	httpClient := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, sha1URL, strings.NewReader(treeURL))
	if err != nil {
		t.Error(err)
	}
	res, err := httpClient.Do(req)
	if err != nil {
		t.Error(err)
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Error(err)
	}

	var parsed map[string]string
	err = json.Unmarshal(bodyBytes, &parsed)
	if err != nil {
		t.Error(err)
	}

	if _, ok := parsed["tree_urls"]; !ok {
		t.Error("Parsed response did not have a top-level 'tree_urls' key")
	}

	if parsed["tree_urls"] != treeURL {
		t.Errorf("Put returned '%s' as the tree URL instead of '%s'", parsed["tree_urls"], treeURL)
	}
}

func TestPostInsert(t *testing.T) {
	sha1 := "60e3da2efd886074e28e44d48cc642f84c25b140"
	treeURL := `[{"label":"tree_0","url":"http://portnoy.iplantcollaborative.org/view/tree/3727f35cc7125567492cab69850f6473"}]`

	mock := NewMockDB()
	n := New(mock)
	server := httptest.NewServer(n.router)
	defer server.Close()

	sha1URL := fmt.Sprintf("%s/%s", server.URL, sha1)
	res, err := http.Post(sha1URL, "", strings.NewReader(treeURL))
	if err != nil {
		t.Error(err)
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Error(err)
	}

	var parsed map[string]string
	err = json.Unmarshal(bodyBytes, &parsed)
	if err != nil {
		t.Error(err)
	}

	if _, ok := parsed["tree_urls"]; !ok {
		t.Error("Parsed response did not have a top-level 'tree_urls' key")
	}

	if parsed["tree_urls"] != treeURL {
		t.Errorf("Post returned '%s' as the tree URL instead of '%s'", parsed["tree_urls"], treeURL)
	}
}

func TestPostUpdate(t *testing.T) {
	sha1 := "60e3da2efd886074e28e44d48cc642f84c25b140"
	treeURL := `[{"label":"tree_0","url":"http://portnoy.iplantcollaborative.org/view/tree/3727f35cc7125567492cab69850f6473"}]`

	mock := NewMockDB()
	if err := mock.insertTreeURLs(sha1, treeURL); err != nil {
		t.Error(err)
	}

	n := New(mock)
	server := httptest.NewServer(n.router)
	defer server.Close()

	sha1URL := fmt.Sprintf("%s/%s", server.URL, sha1)
	res, err := http.Post(sha1URL, "", strings.NewReader(treeURL))
	if err != nil {
		t.Error(err)
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Error(err)
	}

	var parsed map[string]string
	err = json.Unmarshal(bodyBytes, &parsed)
	if err != nil {
		t.Error(err)
	}

	if _, ok := parsed["tree_urls"]; !ok {
		t.Error("Parsed response did not have a top-level 'tree_urls' key")
	}

	if parsed["tree_urls"] != treeURL {
		t.Errorf("Post returned '%s' as the tree URL instead of '%s'", parsed["tree_urls"], treeURL)
	}
}

func TestDelete(t *testing.T) {
	sha1 := "60e3da2efd886074e28e44d48cc642f84c25b140"
	treeURL := `[{"label":"tree_0","url":"http://portnoy.iplantcollaborative.org/view/tree/3727f35cc7125567492cab69850f6473"}]`

	mock := NewMockDB()
	if err := mock.insertTreeURLs(sha1, treeURL); err != nil {
		t.Error(err)
	}

	n := New(mock)
	server := httptest.NewServer(n.router)
	defer server.Close()

	sha1URL := fmt.Sprintf("%s/%s", server.URL, sha1)
	httpClient := &http.Client{}
	req, err := http.NewRequest(http.MethodDelete, sha1URL, nil)
	if err != nil {
		t.Error(err)
	}
	res, err := httpClient.Do(req)
	if err != nil {
		t.Error(err)
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Error(err)
	}

	if len(bodyBytes) > 0 {
		t.Errorf("Delete returned a body when it should not have: %s", string(bodyBytes))
	}

	expectedStatus := http.StatusOK
	actualStatus := res.StatusCode

	if actualStatus != expectedStatus {
		t.Errorf("StatusCode was %d instead of %d", actualStatus, expectedStatus)
	}
}

func TestDeleteUnstored(t *testing.T) {
	sha1 := "60e3da2efd886074e28e44d48cc642f84c25b140"

	mock := NewMockDB()

	n := New(mock)
	server := httptest.NewServer(n.router)
	defer server.Close()

	sha1URL := fmt.Sprintf("%s/%s", server.URL, sha1)
	httpClient := &http.Client{}
	req, err := http.NewRequest(http.MethodDelete, sha1URL, nil)
	if err != nil {
		t.Error(err)
	}
	res, err := httpClient.Do(req)
	if err != nil {
		t.Error(err)
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Error(err)
	}

	if len(bodyBytes) > 0 {
		t.Errorf("Delete returned a body when it should not have: %s", string(bodyBytes))
	}

	expectedStatus := http.StatusOK
	actualStatus := res.StatusCode

	if actualStatus != expectedStatus {
		t.Errorf("StatusCode was %d instead of %d", actualStatus, expectedStatus)
	}
}

func TestFixAddrNoPrefix(t *testing.T) {
	expected := ":70000"
	actual := fixAddr("70000")
	if actual != expected {
		t.Fail()
	}
}

func TestFixAddrWithPrefix(t *testing.T) {
	expected := ":70000"
	actual := fixAddr(":70000")
	if actual != expected {
		t.Fail()
	}
}

func TestNewPostgresDB(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating the mock db: %s", err)
	}
	defer db.Close()

	p := NewPostgresDB(db)
	if p == nil {
		t.Errorf("error from NewPostgresDB(): %s", err)
	}

	if p.db != db {
		t.Error("dbs did not match")
	}
}

func TestHasSHA1(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating the mock db: %s", err)
	}
	defer db.Close()

	p := NewPostgresDB(db)
	if p == nil {
		t.Error("NewPostgresDB() returned nil")
	}

	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM tree_urls WHERE sha1 =").
		WithArgs("sha1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	hasSHA1, err := p.hasSHA1("sha1")
	if err != nil {
		t.Errorf("error from hasSHA1(): %s", err)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations were not met: %s", err)
	}

	if !hasSHA1 {
		t.Error("hasSHA1() returned false")
	}
}

func TestGetTreeURLs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating the mock db: %s", err)
	}
	defer db.Close()

	p := NewPostgresDB(db)
	if p == nil {
		t.Error("NewPostgresDB returned nil")
	}

	mock.ExpectQuery("SELECT tree_urls FROM tree_urls WHERE sha1 =").
		WithArgs("sha1").
		WillReturnRows(sqlmock.NewRows([]string{"tree_urls"}).AddRow("{}"))

	records, err := p.getTreeURLs("sha1")
	if err != nil {
		t.Errorf("error from getTreeURLs(): %s", err)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations were not met: %s", err)
	}

	if len(records) != 1 {
		t.Errorf("number of records returned was %d instead of 1", len(records))
	}

	treeurl := records[0]

	if treeurl != "{}" {
		t.Errorf("tree url was %s instead of '{}'", treeurl)
	}
}

func TestInsertTreeURLs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating the mock db: %s", err)
	}
	defer db.Close()

	p := NewPostgresDB(db)
	if p == nil {
		t.Error("NewPostgresDB returned nil")
	}

	mock.ExpectExec("INSERT INTO tree_urls \\(sha1, tree_urls\\) VALUES").
		WithArgs("sha1", "{}").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err = p.insertTreeURLs("sha1", "{}"); err != nil {
		t.Errorf("error inserting tree urls: %s", err)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations were not met: %s", err)
	}
}

func TestUpdateTreeURLs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating the mock db: %s", err)
	}
	defer db.Close()

	p := NewPostgresDB(db)
	if p == nil {
		t.Error("NewPostgresDB returned nil")
	}

	mock.ExpectExec("UPDATE ONLY tree_urls SET tree_urls =").
		WithArgs("sha1", "{}").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err = p.updateTreeURLs("sha1", "{}"); err != nil {
		t.Errorf("error updating tree urls: %s", err)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations were not met: %s", err)
	}
}

func TestDeleteTreeURLs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating the mock db: %s", err)
	}
	defer db.Close()

	p := NewPostgresDB(db)
	if p == nil {
		t.Error("NewPostgresDB() returned nil")
	}

	mock.ExpectExec("DELETE FROM tree_urls WHERE sha1 =").
		WithArgs("sha1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err = p.deleteTreeURLs("sha1"); err != nil {
		t.Errorf("error deleting tree urls: %s", err)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations were not met: %s", err)
	}
}
