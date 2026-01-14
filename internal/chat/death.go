package chat

import (
  "log"
  "fmt"
  "strings"
  "context"
  "posso-help/internal/db"
  "posso-help/internal/date"
  "posso-help/internal/utils"
  "go.mongodb.org/mongo-driver/bson"
)

type Death struct {
  Phone       string `json:"phone"`
  Name        string `json:"name"`
  Date        string `json:"date"`
  Tag         int64  `json:"tag"`
  Sex         string `json:"sex"`
  Cause       string `json:"cause"`
}

type DeathEntry struct {
  Id       int    `json:"tag"`
  Cause    string `json:"cause"`
}

type DeathMessage struct {
  Date string
  Entries []*DeathEntry
  Total int
}

func (b *DeathMessage) GetCollection() string {
  return "death"
}

func (d *DeathMessage) Parse(message string) bool {
  found := false
  lines := strings.Split(message, "\n")
  for _, line := range lines {
    if date, found := date.ParseAsDateLine(line); found {
      d.Date = date
    }
    if entry := d.parseAsDeathLine(line); entry != nil {
      d.Entries = append(d.Entries, entry)
      d.Total++
      found = true
    }
  }
  return found 
}

func (d *DeathMessage) parseAsDeathLine(line string) (*DeathEntry) {
  var num int
  var cause string
  line = utils.SanitizeLine(line)
	n, err := fmt.Sscanf(line, "%d %s", &num, &cause)
  if err == nil && n == 2 && num > 0 &&
    (utils.StringIsOneOf(cause, DEATHS)) {
      return &DeathEntry{Id:num, Cause:cause}
  }
  return nil
}

func (d *DeathMessage) Text(lang string) string {
  reply := map[string]string {
    "en-US" : "Zap Manejo has detected death data. " +  
              "We added %d deaths. "                 + 
              "To claim your data and see a report " + 
              "sign up at https://dashboard.zapmanejo.com/",
    "pt-BR" : "Zap Manejo detectou dados de óbitos. "     + 
              "Adicionamos %d óbitos. "                   + 
              "Para reivindicar seus dados e visualizar " + 
              "um relatório, cadastre-se em https://dashboard.zapmanejo.com/",
  }

  if lang == "pt-BR" ||  lang == "en-US" {
    return fmt.Sprintf(reply[lang], d.Total)
  }

  log.Printf("Unsupported or Unknown Language: (%s)", lang)
  return fmt.Sprintf(reply["pt-BR"], d.Total)
}

func (d *DeathMessage) Insert(bmv *BaseMessageValues) error {
  collection := db.GetCollection("births")
  log.Printf("updating death message to collection: %v\n", collection)
  for _, death := range d.Entries {
    document := bson.D{bson.E{Key: "cause", Value: death.Cause}}
		filter := bson.M{"tag": death.Id, "account": bmv.Account}
		log.Printf("updating death, filter %+v", filter)
		log.Printf("updating death, document %+v", document)
    result, err := collection.UpdateOne(
			context.TODO(), 
			filter, 
			bson.M{"$set": document},
		)

    if err != nil {
      log.Printf("error inserting death: %v\n", err)
      return err
    }

    log.Printf("update result: %v\n", result)
  }

  log.Printf("death updated successfully")
  return nil
}
