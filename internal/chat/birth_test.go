package chat

import (
  "testing"
  "posso-help/internal/breed"
  "github.com/stretchr/testify/assert"
)

type BirthTest struct {
  Input string
  Found bool
  Birth *BirthEntry
  Area  string
}

// Helper to create a mock BreedParser for testing
func createTestBreedParser() *breed.BreedParser {
  bp := &breed.BreedParser{}
  bp.AddBreed("angus", "angus")
  bp.AddBreed("nelore", "nelore;nel")
  bp.AddBreed("brangus", "brangus")
  bp.AddBreed("cruzado", "cruzado;cruzada")
  bp.AddBreed("murrah", "murrah")
  bp.AddBreed("mediterrâneo", "mediterrâneo")
  bp.AddBreed("jafarabadi", "jafarabadi")
  bp.AddBreed("carabao", "carabao")
  return bp
}

func TestBirthMessageNoArea(t *testing.T) {
  input := `88888 m cruzado`

  bm := &BirthMessage{BreedParser: createTestBreedParser()}
  bm.Parse(input)
  assert.Equal(t, 1, bm.Total, "Total births do not match")
  // Area will be "unknown" when no area is provided but births are found
  assert.NotNil(t, bm.Area, "Area should not be nil")
  assert.Equal(t, "unknown", bm.Area.Name, "Area should be unknown")
}

func TestBirthMessageWithNewArea(t *testing.T) {
  input := `88888 m cruzado
            Jupiter`

  bm := &BirthMessage{BreedParser: createTestBreedParser()}
  bm.Parse(input)
  assert.Equal(t, 1, bm.Total, "Total births do not match")
  assert.NotNil(t, bm.Area, "Area should not be nil")
  assert.Equal(t, "jupiter", bm.Area.Name, "Area does not match")
}


// Test calf parsing with English keyword
func TestCalfMessageEnglish(t *testing.T) {
  input := `calf 12345 f nelore`

  bm := &BirthMessage{BreedParser: createTestBreedParser()}
  found := bm.Parse(input)

  assert.True(t, found, "Should find calf entry")
  assert.Equal(t, 1, bm.Total, "Total should be 1")
  assert.Equal(t, 1, len(bm.Entries), "Should have 1 entry")
  assert.Equal(t, 0, bm.Entries[0].Id, "Tag should be 0 for calf")
  assert.Equal(t, 12345, bm.Entries[0].Dam, "Dam should be 12345")
  assert.Equal(t, "f", bm.Entries[0].Sex, "Sex should be f")
  assert.Equal(t, "nelore", bm.Entries[0].Breed, "Breed should be nelore")
}

// Test calf parsing with Portuguese keyword "bezerro"
func TestCalfMessageBezerro(t *testing.T) {
  input := `bezerro 67890 m angus`

  bm := &BirthMessage{BreedParser: createTestBreedParser()}
  found := bm.Parse(input)

  assert.True(t, found, "Should find calf entry")
  assert.Equal(t, 1, bm.Total, "Total should be 1")
  assert.Equal(t, 0, bm.Entries[0].Id, "Tag should be 0 for calf")
  assert.Equal(t, 67890, bm.Entries[0].Dam, "Dam should be 67890")
  assert.Equal(t, "m", bm.Entries[0].Sex, "Sex should be m")
  assert.Equal(t, "angus", bm.Entries[0].Breed, "Breed should be angus")
}

// Test calf parsing with Portuguese keyword "bezerra"
func TestCalfMessageBezerra(t *testing.T) {
  input := `bezerra 11111 f brangus`

  bm := &BirthMessage{BreedParser: createTestBreedParser()}
  found := bm.Parse(input)

  assert.True(t, found, "Should find calf entry")
  assert.Equal(t, 1, bm.Total, "Total should be 1")
  assert.Equal(t, 0, bm.Entries[0].Id, "Tag should be 0 for calf")
  assert.Equal(t, 11111, bm.Entries[0].Dam, "Dam should be 11111")
  assert.Equal(t, "f", bm.Entries[0].Sex, "Sex should be f")
}

// Test calf parsing with short Portuguese keyword "bez"
func TestCalfMessageBez(t *testing.T) {
  input := `bez 22222 m nelore`

  bm := &BirthMessage{BreedParser: createTestBreedParser()}
  found := bm.Parse(input)

  assert.True(t, found, "Should find calf entry")
  assert.Equal(t, 1, bm.Total, "Total should be 1")
  assert.Equal(t, 0, bm.Entries[0].Id, "Tag should be 0 for calf")
  assert.Equal(t, 22222, bm.Entries[0].Dam, "Dam should be 22222")
}

// Test mixed births and calves in same message
func TestMixedBirthsAndCalves(t *testing.T) {
  input := `88888 m angus
calf 12345 f nelore
bezerro 12345 m nelore`

  bm := &BirthMessage{BreedParser: createTestBreedParser()}
  found := bm.Parse(input)

  assert.True(t, found, "Should find entries")
  assert.Equal(t, 3, bm.Total, "Total should be 3")
  assert.Equal(t, 3, len(bm.Entries), "Should have 3 entries")

  // First entry: regular birth
  assert.Equal(t, 88888, bm.Entries[0].Id, "First entry tag should be 88888")
  assert.Equal(t, 0, bm.Entries[0].Dam, "First entry dam should be 0")

  // Second entry: calf
  assert.Equal(t, 0, bm.Entries[1].Id, "Second entry tag should be 0")
  assert.Equal(t, 12345, bm.Entries[1].Dam, "Second entry dam should be 12345")

  // Third entry: bezerro
  assert.Equal(t, 0, bm.Entries[2].Id, "Third entry tag should be 0")
  assert.Equal(t, 12345, bm.Entries[2].Dam, "Third entry dam should be 12345")
}

// Test calf with area
func TestCalfMessageWithArea(t *testing.T) {
  input := `calf 12345 f nelore
fazenda norte`

  bm := &BirthMessage{BreedParser: createTestBreedParser()}
  found := bm.Parse(input)

  assert.True(t, found, "Should find calf entry")
  assert.Equal(t, 1, bm.Total, "Total should be 1")
  assert.NotNil(t, bm.Area, "Area should not be nil")
  assert.Equal(t, "fazenda norte", bm.Area.Name, "Area should be fazenda norte")
}

// Test that invalid calf lines are not parsed
func TestInvalidCalfLines(t *testing.T) {
  // Missing dam number
  bm1 := &BirthMessage{BreedParser: createTestBreedParser()}
  assert.False(t, bm1.Parse("calf f nelore"), "Should not parse calf without dam")

  // Invalid sex
  bm2 := &BirthMessage{BreedParser: createTestBreedParser()}
  assert.False(t, bm2.Parse("calf 12345 x nelore"), "Should not parse calf with invalid sex")

  // Dam = 0 should not be valid
  bm3 := &BirthMessage{BreedParser: createTestBreedParser()}
  assert.False(t, bm3.Parse("calf 0 f nelore"), "Should not parse calf with dam=0")
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
