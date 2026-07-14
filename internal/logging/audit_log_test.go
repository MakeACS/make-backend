package logging

import (
	"make-backend/internal/database/models"
	"testing"
)

func TestPlainStringGenerationOrder(t *testing.T) {
	user1 := models.User{
		Id:        0,
		Firstname: "John",
		Lastname:  "Doe",
	}
	user2 := models.User{
		Id:        2,
		Firstname: "Jane",
		Lastname:  "Doe",
	}
	str := CreatePlainString("{user} elevated {user} to admin privileges", user1.LogEntity(), user2.LogEntity())
	ref := "John Doe elevated Jane Doe to admin privileges"
	if str != ref {
		t.Errorf("reference and created string don't agree. Wanted `%s` got `%s`", str, ref)
	}
}

func TestFormattedStringGenerationOrder(t *testing.T) {
	user1 := models.User{
		Id:        0,
		Firstname: "John",
		Lastname:  "Doe",
	}
	user2 := models.User{
		Id:        2,
		Firstname: "Jane",
		Lastname:  "Doe",
	}
	str := CreateFormatString("{user} elevated {user} to admin privileges", user1.LogEntity(), user2.LogEntity())
	ref := "{user:0:John Doe} elevated {user:2:Jane Doe} to admin privileges"
	if str != ref {
		t.Errorf("reference and created string don't agree. Wanted `%s` got `%s`", str, ref)
	}
}
