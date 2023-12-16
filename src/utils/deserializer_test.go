package utils

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
	"wild_project/src/models"
)

func TestValidateOrder(t *testing.T) {
	testCases := []struct {
		name    string
		order   models.Order
		withErr bool
	}{
		{
			name:    "Valid Order",
			order:   models.Order{OrderUID: "123", TrackNumber: "123"},
			withErr: false,
		},
		{
			name:    "Valid Order",
			order:   models.Order{OrderUID: "", TrackNumber: "123"},
			withErr: true,
		},
		{
			name:    "Valid Order",
			order:   models.Order{OrderUID: "123", TrackNumber: ""},
			withErr: true,
		},
	}

	// запускаем
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateOrder(&tc.order)
			if tc.withErr && err == nil {
				t.Errorf("%s: ожидалась ошибка, но она не возникла", tc.name)
			} else if !tc.withErr && err != nil {
				t.Errorf("%s: не ожидалась ошибка, но возникла: %v", tc.name, err)
			}

		})
	}
}

func TestDeserializeOrder(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		name      string
		jsonOrder string
		wantOrder models.Order
		wantErr   bool
	}{
		{
			name:      "Valid JSON",
			jsonOrder: `{"OrderUID": "123", "TrackNumber": "ABC123"}`,
			wantOrder: models.Order{OrderUID: "123", TrackNumber: "ABC123"},
			wantErr:   false,
		},
		{
			name:      "Invalid JSON",
			jsonOrder: `{"OrderUID": 123, "TrackNumber": "ABC123"}`, // Некорректный JSON для Order
			wantErr:   true,
		},
		{
			name:      "Empty JSON",
			jsonOrder: `{}`,
			wantOrder: models.Order{}, // Пустой Order
			wantErr:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotOrder, err := DeserializeOrder(tc.jsonOrder)

			if tc.wantErr {
				assert.Error(err, "%s: ожидалась ошибка", tc.name)
			} else {
				assert.NoError(err, "%s: не ожидалась ошибка", tc.name)
				assert.True(
					jsonEqual(
						gotOrder, tc.wantOrder),
					"%s: полученный результат %v, ожидался %v", tc.name, gotOrder, tc.wantOrder)
			}
		})
	}
}

// jsonEqual сравнивает два объекта Order на равенство после их сериализации в JSON.
func jsonEqual(a, b models.Order) bool {
	aJSON, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bJSON, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(aJSON) == string(bJSON)
}
