package chat

import (
  "testing"
  "github.com/stretchr/testify/assert"
)

type DeathTest struct {
  Input string
  Found bool
  Death *DeathEntry
}

func TestDeathMessage(t *testing.T) {
  input := `2235 natimorto
            2236 Aborto`

  dm := &DeathMessage{}
  dm.Parse(input)
  assert.Equal(t, 2, dm.Total, "Total deaths do not match")
}

func TestParseAsDeathLine(t *testing.T) {
  dm := &DeathMessage{}
  tests := []DeathTest {
    DeathTest{"2235 natimorto", true, &DeathEntry{2235, NATIMORTO}},
    DeathTest{"2236 Aborto",    true, &DeathEntry{2236, ABORTO}},
    DeathTest{"1225 Morreu",    true, &DeathEntry{1225, MORREU}},
    DeathTest{"1226 Morto",     true, &DeathEntry{1226, MORTO}},
  }

  for index, test := range tests {
    death := dm.parseAsDeathLine(test.Input)

    if (death == nil && !test.Found) {
      // Success, expected nothing back and got nil back
      continue
    }

    if (death == nil && test.Found) {
      t.Errorf("TestParseAsDeath() expected Death but got nil %d", index)
      continue
    }

    if death.Id != test.Death.Id {
      t.Errorf("TestParseAsDeath() Id Mismatch index: [%d] expected [%d] got: [%d]",
        index, test.Death.Id, death.Id)
    }

    if death.Cause != test.Death.Cause {
      t.Errorf("TestParseAsDeath() Cause Mismatch index: [%d] expected [%s] got: [%s]",
        index, test.Death.Cause, death.Cause)
    }
  }
}
