package matchengine

import (
	"encoding/json"
	"testing"
)

func TestImageCropValidation(t *testing.T) {
	valid := ImageCrop{X: 0.2, Y: 0.08, Width: 0.6, Height: 0.56}
	if !valid.Valid() {
		t.Fatal("expected seeded 2:3 to 5:7 crop to be valid")
	}
	for _, invalid := range []ImageCrop{
		{X: -0.1, Y: 0, Width: 0.6, Height: 0.56},
		{X: 0.8, Y: 0, Width: 0.6, Height: 0.56},
		{X: 0, Y: 0, Width: 0.6, Height: 0.2},
		{},
	} {
		if invalid.Valid() {
			t.Fatalf("expected invalid crop: %+v", invalid)
		}
	}
}

func TestPlayerStateVisualJSONIsOptional(t *testing.T) {
	plain, err := json.Marshal(PlayerState{PlayerID: "p"})
	if err != nil {
		t.Fatal(err)
	}
	if string(plain) == "" || containsJSONKey(plain, "card_image") || containsJSONKey(plain, "avatar_crop") {
		t.Fatalf("empty visual fields should be omitted: %s", plain)
	}
	crop := &ImageCrop{X: 0.2, Y: 0.08, Width: 0.6, Height: 0.56}
	visual, err := json.Marshal(PlayerState{PlayerID: "p", CardImage: "player-cards/p.png", AvatarCrop: crop})
	if err != nil {
		t.Fatal(err)
	}
	if !containsJSONKey(visual, "card_image") || !containsJSONKey(visual, "avatar_crop") {
		t.Fatalf("visual fields missing from JSON: %s", visual)
	}
}

func containsJSONKey(data []byte, key string) bool {
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		return false
	}
	_, ok := decoded[key]
	return ok
}
