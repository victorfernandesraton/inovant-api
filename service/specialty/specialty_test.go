package specialty

import (
	"testing"

	"github.com/jmoiron/sqlx"
	m "gitlab.com/falqon/inovantapp/backend/models"
)

var psqlInfo = ("host=localhost port=5432 user=postgres password=123 dbname=inovant_test sslmode=disable")

func TestCreateSpecialty(t *testing.T) {
	db, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	sSpecialty := m.Specialty{
		Name:        "Nutrologia",
		Description: "Estudo de nutrição",
	}

	specCre := Creator{DB: db}
	cre, err := specCre.Run(&sSpecialty)
	if err != nil {
		t.Errorf("Create Specialty failed, expected struct got %v ", err)
	} else {
		t.Logf("Create Specialty success, expected %v, got %v", &sSpecialty, cre)
	}
}

func TestUpdateSpecialty(t *testing.T) {
	db, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	sSpecialty := m.Specialty{
		SpecID:      26,
		Name:        "Teste",
		Description: "Teste",
	}

	specUpd := Updater{DB: db}
	upd, err := specUpd.Run(&sSpecialty)
	if err != nil {
		t.Errorf("Update Specialty failed, expected struct got %v ", err)
	} else {
		t.Logf("Update Specialty success, expected %v, got %v", &sSpecialty, upd)
	}
}

func TestDeleteSpecialty(t *testing.T) {
	db, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	sSpecialty := int64(26)

	specDel := Deleter{DB: db}
	del, err := specDel.Run(sSpecialty)
	if err != nil {
		t.Errorf("Delete Specialty failed, expected struct got %v ", err)
	} else {
		t.Logf("Delete Specialty success, expected %v, got %v", del, del)
	}
}

func TestGetSpecialty(t *testing.T) {
	db, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	sSpecialty := int64(22)

	specGet := Getter{DB: db}
	get, err := specGet.Run(sSpecialty)
	if err != nil {
		t.Errorf("Get Specialty failed, expected struct got %v ", err)
	} else {
		t.Logf("Get Specialty success, expected %v, got %v", get, get)
	}
}

func TestListSpecialty(t *testing.T) {
	db, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	name := "nutri"
	description := "tri"
	limit := int64(10)
	offset := int64(0)
	sSpecialty := m.FilterSpecialty{
		Name:        &name,
		Description: &description,
		Limit:       &limit,
		Offset:      &offset,
	}

	specList := Lister{DB: db}
	list, err := specList.Run(sSpecialty)
	if err != nil {
		t.Errorf("List Specialty failed, expected struct got %v ", err)
	} else {
		t.Logf("List Specialty success, expected %v, got %v", list, list)
	}
}
