package logging

import (
	"make-backend/internal/database/models"
	"testing"
)

var user1 = models.User{
	Id:        0,
	Firstname: "John",
	Lastname:  "Doe",
}
var user2 = models.User{
	Id:        2,
	Firstname: "Jane",
	Lastname:  "Doe",
}
var equipment1 = models.Equipment{
	Id:   1,
	Name: "Bandsaw",
}

func TestPlainStringGenerationOrder(t *testing.T) {

	str := CreatePlainString("{user} elevated {user} to admin privileges", user1.LogEntity(), user2.LogEntity())
	ref := "John Doe elevated Jane Doe to admin privileges"
	if str != ref {
		t.Errorf("reference and created string don't agree. Wanted `%s` got `%s`", str, ref)
	}
}

func TestFormattedStringGenerationOrder(t *testing.T) {
	str := CreateFormatString("{user} elevated {user} to admin privileges", user1.LogEntity(), user2.LogEntity())
	ref := "{user:0:John Doe} elevated {user:2:Jane Doe} to admin privileges"
	if str != ref {
		t.Errorf("reference and created string don't agree. Wanted `%s` got `%s`", str, ref)
	}
}

func TestJsonDataGenerationMultipleDifferent(t *testing.T) {
	data := dataForEntities("{user} created {equipment}", []models.LogEntity{user1.LogEntity(), equipment1.LogEntity()})

	if len(data) != 4 {
		t.Errorf("only expected 4 pieces of data")
	}
	if _, ok := data["user_id"]; !ok {
		t.Errorf("expected data for user_id")
	}
	if _, ok := data["user_label"]; !ok {
		t.Errorf("expected data for user_label")
	}

	if _, ok := data["equipment_id"]; !ok {
		t.Errorf("expected data for equipment_id")
	}
	if _, ok := data["equipment_label"]; !ok {
		t.Errorf("expected data for equipment_label")
	}

}

func TestJsonDataGenerationMultipleSame(t *testing.T) {
	data := dataForEntities("{user} elevated {user} to admin privileges", []models.LogEntity{user1.LogEntity(), user2.LogEntity()})

	if len(data) != 4 {
		t.Errorf("only expected 4 pieces of data")
	}
	if _, ok := data["user_0_id"]; !ok {
		t.Errorf("expected data for user_0_id")
	}
	if _, ok := data["user_0_label"]; !ok {
		t.Errorf("expected data for user_0_label")
	}

	if _, ok := data["user_1_id"]; !ok {
		t.Errorf("expected data for user_1_id")
	}
	if _, ok := data["user_1_label"]; !ok {
		t.Errorf("expected data for user_1_label")
	}

}
