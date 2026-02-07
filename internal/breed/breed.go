package breed

import (
	"context"
	"log"
	"strings"

	"posso-help/internal/db"
	"posso-help/internal/utils"

	"go.mongodb.org/mongo-driver/bson"
)

type Breed struct {
	Name    string `bson:"name"`
	Matches string `bson:"matches"`
}

type BreedParser struct {
	breeds []*Breed
}

// LoadBreedsByAccount loads breeds for the given account plus global breeds
func (bp *BreedParser) LoadBreedsByAccount(account string) error {
	collection := db.GetCollection("breeds")

	// Include both account-specific breeds and global breeds (all zeros account)
	accounts := []string{account, "000000000000000000000000"}
	filter := bson.M{"account": bson.M{"$in": accounts}}

	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		log.Printf("Error reading breeds for account: %v", account)
		return err
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		breed := &Breed{}
		if err := cursor.Decode(breed); err != nil {
			log.Printf("Error decoding breed document: %v", err)
			continue
		}
		log.Printf("LoadBreedsByAccount(%s): %s  %s", account, breed.Name, breed.Matches)
		bp.breeds = append(bp.breeds, breed)
	}

	return cursor.Err()
}

// MatchBreed checks if the given text matches any breed's matches
// Returns the breed name if found, empty string otherwise
func (bp *BreedParser) MatchBreed(text string) (string, bool) {
	text = strings.ToLower(strings.TrimSpace(text))
	for _, breed := range bp.breeds {
		matches := utils.SplitAndTrim(strings.ToLower(breed.Matches))
		if utils.StringIsOneOf(text, matches) {
			log.Printf("MatchBreed: found, name=%s for text=%s", breed.Name, text)
			return breed.Name, true
		}
	}
	return "", false
}

// IsValidBreed checks if the given breed name matches any loaded breed
func (bp *BreedParser) IsValidBreed(name string) bool {
	_, found := bp.MatchBreed(name)
	return found
}

// GetBreedNames returns a slice of all loaded breed names
func (bp *BreedParser) GetBreedNames() []string {
	names := make([]string, len(bp.breeds))
	for i, breed := range bp.breeds {
		names[i] = breed.Name
	}
	return names
}
