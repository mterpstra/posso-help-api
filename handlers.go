package main
import (
  "os"
  "io"
  "log"
  "fmt"
  "net/http"
  "encoding/csv"
  "encoding/json"
  "context"
  "strconv"
  "strings"
  "posso-help/internal/chat"
  "posso-help/internal/db"
  "posso-help/internal/user"
  "github.com/gorilla/mux"

  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson/primitive"
)

func HandleDownload(w http.ResponseWriter, r *http.Request) {
  log.Printf("HandleDownload()\n")
  vars := mux.Vars(r)
  datatype := vars["datatype"]

  ctx := r.Context()
  userID := ctx.Value("user_id")
  if userID == nil {
    log.Printf("could not get userid from context")
    http.Error(w, "Authorization header required", http.StatusUnauthorized)
    return
  }

  user, err := user.Read(userID.(string))
  if err != nil {
    log.Printf("could not read userID from context")
    http.Error(w, "User Not Found", http.StatusNotFound)
    return
  }

  data, err := db.ReadOrdered(datatype, user.Account)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest) 
    fmt.Fprintf(w, "%v", err)
    return 
  }

  csv, err := ConvertBsonToCsv(data) 
  if err != nil {
    w.WriteHeader(http.StatusBadRequest) 
    fmt.Fprintf(w, "%v", err)
    return 
  }

  length := strconv.Itoa(len(csv))
  disposition := fmt.Sprintf("attachment; filename=\"%s.csv\"", datatype)
  w.Header().Add("Content-Type", "text/html")
  w.Header().Add("Content-Length", length)
  w.Header().Add("Content-Disposition", disposition)
  fmt.Fprint(w, string(csv))

  return 
}

func HandleUpload(w http.ResponseWriter, r *http.Request) {
  log.Printf("Handle Upload")

  vars := mux.Vars(r)
  datatype := vars["datatype"]
  log.Printf("HandleUpload: %s", datatype)
  ctx := r.Context()
  userID := ctx.Value("user_id")
  if userID == nil {
    log.Printf("could not get userid from context")
    http.Error(w, "Authorization header required", http.StatusUnauthorized)
    return
  }
  u, err := user.Read(userID.(string))
  if err != nil {
    log.Printf("could not read userID from context")
    http.Error(w, "User Not Found", http.StatusNotFound)
    return
  }

  // The name "csvfile" should match the 'name' attribute in the HTML form's input tag.
  file, handler, err := r.FormFile("csvFile")
  if err != nil {
    http.Error(w, "Error retrieving the file", http.StatusBadRequest)
    log.Println(err)
    return
  }
  defer file.Close()

  log.Printf("Uploaded File: %+v\n", handler.Filename)
  log.Printf("File Size: %+v\n", handler.Size)
  log.Printf("MIME Header: %+v\n", handler.Header)

  csvReader := csv.NewReader(file)
  headers, err := csvReader.Read()
  if err != nil {
    http.Error(w, "Error reading CSV header", http.StatusInternalServerError)
    log.Println(err)
    return
  }
  fieldCount := len(headers)
  log.Printf("Field Count: %+v\n", fieldCount)
  record := bson.M{"account": u.Account}
  collection := db.GetCollection(datatype);
  for {
    row, err := csvReader.Read()
    if err == io.EOF {
      break 
    }
    if err != nil {
      fmt.Printf("Error reading row: %v\n", err)
      return
    }

    for i, header := range headers {
      if i < len(row) { // Ensure index is within bounds of the row
        key := strings.TrimSpace(header)
        value := strings.TrimSpace(row[i])
        record[key] = value

        // Hack for tag being an int values
        if key == "tag" || key == "amount" || key == "temperature" {
          tag, err := strconv.Atoi(value)
          if err == nil {
            record[key] = tag 
          }
        }
      }
    }
    log.Printf("record: %+v\n", record)
    result, err := collection.InsertOne(context.TODO(), record)
    if err != nil {
      log.Printf("Error inserting %s record: %+v, err: %v", 
      datatype, record, err)
    }
    log.Printf("result: %+v\n", result)

  }
}

func HandleUploadJSON(w http.ResponseWriter, r *http.Request) {
  log.Printf("HandleUploadJSON")

  vars := mux.Vars(r)
  datatype := vars["datatype"]
  log.Printf("HandleUploadJSON: %s", datatype)

  ctx := r.Context()
  userID := ctx.Value("user_id")
  if userID == nil {
    log.Printf("could not get userid from context")
    http.Error(w, "Authorization header required", http.StatusUnauthorized)
    return
  }
  u, err := user.Read(userID.(string))
  if err != nil {
    log.Printf("could not read userID from context")
    http.Error(w, "User Not Found", http.StatusNotFound)
    return
  }

  defer r.Body.Close()
  bodyBytes, err := io.ReadAll(r.Body)
  if err != nil {
    http.Error(w, "Error reading request body", http.StatusInternalServerError)
    log.Printf("Error reading request body: %v", err)
    return
  }

  var records []map[string]interface{}
  err = json.Unmarshal(bodyBytes, &records)
  if err != nil {
    http.Error(w, "Error parsing JSON", http.StatusBadRequest)
    log.Printf("Error parsing JSON: %v", err)
    return
  }

  collection := db.GetCollection(datatype)
  inserted := 0
  for _, record := range records {
    record["account"] = u.Account

    // Type coercion for known integer fields
    for _, key := range []string{"tag", "amount", "temperature"} {
      if val, ok := record[key]; ok {
        switch v := val.(type) {
        case float64:
          record[key] = int(v)
        case string:
          if n, err := strconv.Atoi(v); err == nil {
            record[key] = n
          }
        }
      }
    }

    log.Printf("record: %+v", record)
    result, err := collection.InsertOne(context.TODO(), record)
    if err != nil {
      log.Printf("Error inserting %s record: %+v, err: %v", datatype, record, err)
      continue
    }
    log.Printf("result: %+v", result)
    inserted++
  }

  w.Header().Set("Content-Type", "application/json")
  fmt.Fprintf(w, `{"inserted":%d,"total":%d}`, inserted, len(records))
}

func HandleDataGet(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  datatype := vars["datatype"]

  ctx := r.Context()
  userID := ctx.Value("user_id")
  if userID == nil {
    log.Printf("could not get userid from context")
    http.Error(w, "Authorization header required", http.StatusUnauthorized)
    return
  }

  user, err := user.Read(userID.(string))
  if err != nil {
    log.Printf("could not read userID from context")
    http.Error(w, "User Not Found", http.StatusNotFound)
    return
  }

  filters := map[string]string{}
  searchFields := r.URL.Query().Get("search_fields")
  searchValues := r.URL.Query().Get("search_values")
  if searchFields != "" && searchValues != "" {
    fields := strings.Split(searchFields, ",")
    values := strings.Split(searchValues, ",")
    if len(fields) != len(values) {
      http.Error(w, "search_fields and search_values must have the same number of items", http.StatusBadRequest)
      return
    }
    for i, field := range fields {
      filters[field] = values[i]
    }
  }

  data, err := db.ReadUnordered(datatype, user.Account, filters)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    fmt.Fprintf(w, "%v", err)
    return
  }

  json, err := json.Marshal(data)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest) 
    fmt.Fprintf(w, "%v", err)
    return 
  }
  fmt.Fprint(w, string(json))
  return 
}

func HandleDataPut(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  datatype := vars["datatype"]
  log.Printf("HandleDataPut: %s", datatype)
  ctx := r.Context()
  userID := ctx.Value("user_id")
  if userID == nil {
    log.Printf("could not get userid from context")
    http.Error(w, "Authorization header required", http.StatusUnauthorized)
    return
  }
  u, err := user.Read(userID.(string))
  if err != nil {
    log.Printf("could not read userID from context")
    http.Error(w, "User Not Found", http.StatusNotFound)
    return
  }
  collection := db.GetCollection(datatype);
  defer r.Body.Close()
  bodyBytes, err := io.ReadAll(r.Body)
  if err != nil {
    http.Error(w, "Error reading request body", http.StatusInternalServerError)
    log.Printf("Error reading request body: %v", err)
    return
  }

  log.Printf("user: %s  collection: %v  body: %s",
		u.Email, collection, string(bodyBytes))

  data := make(map[string]interface{})
  err = json.Unmarshal(bodyBytes, &data)
  if err != nil {
    http.Error(w, "Error unmarshalling JSON", http.StatusBadRequest)
    log.Printf("Error unmarshalling JSON: %v", err)
    return 
  }

  objID, err := primitive.ObjectIDFromHex(data["_id"].(string))
  if err != nil {
    http.Error(w, "Error _id required", http.StatusBadRequest)
    log.Printf("Error _id required: %v", err)
    return 
  }
  filter := bson.M{"_id": objID, "account": u.Account}
  delete(data, "_id")

  _, err = collection.UpdateOne(context.TODO(), filter, bson.M{"$set": data})
  if err != nil {
    http.Error(w, "Error Updating Data", http.StatusBadRequest)
    log.Printf("Error Updating Data: %v", err)
    return 
  }

  log.Printf("Successful Update")
  return
}

func HandleDataPatch(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  datatype := vars["datatype"]
  log.Printf("HandleDataPatch: %s", datatype)
  ctx := r.Context()
  userID := ctx.Value("user_id")
  if userID == nil {
    log.Printf("could not get userid from context")
    http.Error(w, "Authorization header required", http.StatusUnauthorized)
    return
  }
  u, err := user.Read(userID.(string))
  if err != nil {
    log.Printf("could not read userID from context")
    http.Error(w, "User Not Found", http.StatusNotFound)
    return
  }
  collection := db.GetCollection(datatype);
  defer r.Body.Close()
  bodyBytes, err := io.ReadAll(r.Body)
  if err != nil {
    http.Error(w, "Error reading request body", http.StatusInternalServerError)
    log.Printf("Error reading request body: %v", err)
    return
  }
  log.Printf("user: %s  collection: %v  body: %s",
		u.Email, collection, string(bodyBytes))

  data := make(map[string]interface{})
  err = json.Unmarshal(bodyBytes, &data)
  if err != nil {
    http.Error(w, "Error unmarshalling JSON", http.StatusBadRequest)
    log.Printf("Error unmarshalling JSON: %v", err)
    return 
  }

  objID, err := primitive.ObjectIDFromHex(data["_id"].(string))
  if err != nil {
    http.Error(w, "Error _id required", http.StatusBadRequest)
    log.Printf("Error _id required: %v", err)
    return 
  }
  filter := bson.M{"_id": objID, "account": u.Account}
  delete(data, "_id")

  _, err = collection.UpdateOne(context.TODO(), filter, bson.M{"$set": data})
  if err != nil {
    http.Error(w, "Error Updating Data", http.StatusBadRequest)
    log.Printf("Error Updating Data: %v", err)
    return 
  }


  // If we updated the user, we want to give a new JWT back as well.
  if (datatype == "users") {
    log.Printf("Update to user, generating new JWT");
    updatedUser, err := user.Read(userID.(string))
    if err != nil {
      log.Printf("Error reading user to get new JWT")
      return
    }

    jsonData, err := json.Marshal(updatedUser)
    if err != nil {
      log.Printf("Error marshaling updated user %+v", err)
      return
    }
    w.Header().Set("X-New-User", string(jsonData))
  }

  return
}

func HandleDataPost(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  datatype := vars["datatype"]
  // @todo: check valid values for datatype 
  log.Printf("HandleDataPost: %s", datatype)
  ctx := r.Context()
  userID := ctx.Value("user_id")
  if userID == nil {
    log.Printf("could not get userid from context")
    http.Error(w, "Authorization header required", http.StatusUnauthorized)
    return
  }

  u, err := user.Read(userID.(string))
  if err != nil {
    log.Printf("could not read userID from context")
    http.Error(w, "User Not Found", http.StatusNotFound)
    return
  }

  collection := db.GetCollection(datatype);

  defer r.Body.Close()
  bodyBytes, err := io.ReadAll(r.Body)
  if err != nil {
    http.Error(w, "Error reading request body", http.StatusInternalServerError)
    log.Printf("Error reading request body: %v", err)
    return
  }

  log.Printf("user: %s  collection: %v  body: %s",
		u.Email, collection, string(bodyBytes))

  data := make(map[string]interface{})

  err = json.Unmarshal(bodyBytes, &data)
  if err != nil {
    http.Error(w, "Error unmarshalling JSON", http.StatusBadRequest)
    log.Printf("Error unmarshalling JSON: %v", err)
    return 
  }

  data["account"] = u.Account

  // Name was used before
  data["created_by"] = u.GetDisplayName()

  _, err = collection.InsertOne(context.TODO(), data)
  if err != nil {

    if mongo.IsDuplicateKeyError(err) {
      http.Error(w, "duplicate_key_error", http.StatusBadRequest)
    } else {
      http.Error(w, "Error Inserting Data", http.StatusBadRequest)
    }
    log.Printf("Error Inserting Data: %v", err)
    return 
  }

  return 
}

func HandleDataDelete(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  datatype := vars["datatype"]
  id := vars["id"]
  collection := db.GetCollection(datatype);
  log.Printf("Deleteing:  datatype: %s, id: %s", datatype, id)


  objID, err := primitive.ObjectIDFromHex(id)
  if err != nil {
    http.Error(w, "invalid_id", http.StatusBadRequest)
    log.Printf("invalid id: %s %v", id, err)
    return
  }

  filter := bson.M{"_id": objID}
  deleteResult, err := collection.DeleteOne(context.TODO(), filter)
  if err != nil {
    log.Fatal(err)
    http.Error(w, "error_deleting_id", http.StatusBadRequest)
    log.Printf("Error Deleting ID: %s %v", id, err)
    return
  }
  log.Printf("Number of records deleted: %d by filter %v", 
  deleteResult.DeletedCount, filter)
  return 
}

func HandleChatMessage(w http.ResponseWriter, r *http.Request) {
  log.Printf("HandleChatMessage")
  defer r.Body.Close()
  bodyBytes, err := io.ReadAll(r.Body)
  if err != nil {
    http.Error(w, "Error reading request body", http.StatusInternalServerError)
    log.Printf("could not read body: %v\n", err)
    return
  }
  log.Printf("ChatMessage: %s\n", string(bodyBytes))

  chatMessage := &chat.ChatMessage{}
  err = json.Unmarshal(bodyBytes, chatMessage)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Printf("unmarshal error: %v\n", err)
    return
  }

  err = chat.ProcessEntries(chatMessage.Entries)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Printf("error processing entries: %v\n", err)
    return
  }

  log.Printf("success handing chat message")
}

func HandleHubChallenge(w http.ResponseWriter, r *http.Request) {
  log.Printf("HandleHubChallenge")
  osToken := os.Getenv("HUB_TOKEN")
  mode := r.URL.Query().Get("hub.mode")
  token := r.URL.Query().Get("hub.verify_token")
  challenge := r.URL.Query().Get("hub.challenge")

  if len(osToken) < 1 {
    http.Error(w, "environment_error", http.StatusBadRequest)
    return
  }

  if mode != "subscribe" {
    http.Error(w, "invalid_mode", http.StatusBadRequest)
    return
  }

  if token != osToken {
    http.Error(w, "invalid_token", http.StatusBadRequest)
    return
  }

  fmt.Fprint(w, challenge)
}
