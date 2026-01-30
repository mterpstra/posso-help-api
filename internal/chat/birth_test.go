package chat

import (
  "testing"
  "github.com/stretchr/testify/assert"
)

type BirthTest struct {
  Input string
  Found bool
  Birth *BirthEntry
  Area  string
}

func TestBirthMessageNoArea(t *testing.T) {
  input := `88888 m Cruzado`

  bm := &BirthMessage{}
  bm.Parse(input)
  assert.Equal(t, bm.Total, 1, "Total births do not match")
  assert.Nil(t, bm.Area, "Area should be nil")
}

func TestBirthMessageWithArea(t *testing.T) {
  input := `88888 m Cruzado
            espirito santo`

  bm := &BirthMessage{}
  bm.Parse(input)
  assert.Equal(t, bm.Total, 1, "Total births do not match")
  assert.Equal(t, bm.Area.Name, "espirito santo", "Area does not match")
}

func TestBirthMessageWithNewArea(t *testing.T) {
  input := `88888 m Cruzado
            Jupiter`

  bm := &BirthMessage{}
  bm.Parse(input)
  assert.Equal(t, 1, bm.Total, "Total births do not match")
  assert.NotNil(t, bm.Area, "Area should not be nil")
  assert.Equal(t, "jupiter", bm.Area.Name, "Area does not match")
}


func TestParseBithLine(t *testing.T) {
  /* @todo: this test needs to be revsited
  bm := &BirthMessage{}
  tests := []BirthTest {
    //   INPUT               parsed   Birth{id,    sex,    breed}     area

    // Success Birth Cases
    BirthTest{"1111 m angus",      true,  &BirthEntry{1111,  MALE,   ANGUS},     ""},
    BirthTest{"1212 F   nelore",   true,  &BirthEntry{1212,  FEMALE, NELORE},    ""},
    BirthTest{"342   f Brangus",   true,  &BirthEntry{342,   FEMALE, BRANGUS},   ""},
    BirthTest{"  99 M Sta. Zelia", true,  &BirthEntry{99,    MALE,   STA_ZELIA}, ""},
    BirthTest{"88888 m Cruzado  ", true,  &BirthEntry{88888, MALE,   CRUZADO},   ""},

    // Just some text, should all be ignored
    BirthTest{"Nothing parsed",    false, nil, ""},
    BirthTest{"",                  false, nil, ""},
    BirthTest{"just \n stime \n",  false, nil, ""},

    // Birth and Area on the same line
    BirthTest{"12941 F Angus Filhos de Eva",  true,  &BirthEntry{12941,   FEMALE,   ANGUS}, "filhos de eva"},
  }

  for index, test := range tests {
    birth := bm.parseAsBirthLine(test.Input)

    // Not supposed to be found, this is good
    if !test.Found && birth == nil {
      continue
    }

    if test.Found && birth == nil {
      t.Errorf("ParseLineAsBirth() Failed [%d, %v]", index, test.Input)
      continue
    }

    assert.Equal(t, birth.Id, test.Birth.Id, "Birth Id does not match")
    assert.Equal(t, birth.Sex, test.Birth.Sex, "Birth Sex does not match")
    assert.Equal(t, birth.Breed, test.Birth.Breed, "Birth Breed does not match")
    assert.Equal(t, area, test.Area, "Birth Area does not match")
  }
  */
}
