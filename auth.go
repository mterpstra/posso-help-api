package main

import (
  "context"
  "math/rand"
  "encoding/json"
  "errors"
  "log"
  "fmt"
  "net/http"
  "os"
  "strings"
  "time"
  "regexp"

  "github.com/golang-jwt/jwt/v5"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/bson/primitive"

  "posso-help/internal/email"
  "posso-help/internal/password"
  "posso-help/internal/user"
  "posso-help/internal/db"
)

// JWT Claims structure
type Claims struct {
  UserID      string `json:"user_id"`
  Email       string `json:"email"`
  PhoneNumber string `json:"phone_number"`
  IsActive    bool   `json:"is_active"`
  jwt.RegisteredClaims
}

// Registration request structure
type RegisterRequest struct {
  Name        string `json:"name,omitempty"`
  Email       string `json:"email"`
  Password    string `json:"password"`
  PhoneNumber string `json:"phone_number"`
}

// Login request structure
type LoginRequest struct {
  Email    string `json:"email"`
  Password string `json:"password"`
}

// Auth response structure
type AuthResponse struct {
  Success      bool   `json:"success"`
  Message      string `json:"message"`
  Token        string `json:"token,omitempty"`
  User    *user.User  `json:"user,omitempty"`
  VerificationCode string `json:"verification_code,omitempty"`
}

// Email verification structure
type EmailVerification struct {
  ID          primitive.ObjectID `bson:"_id,omitempty"`
  UserID      primitive.ObjectID `bson:"user_id"`
  Email       string             `bson:"email"`
  Code        string             `bson:"code"`
  ExpiresAt   time.Time          `bson:"expires_at"`
  CreatedAt   time.Time          `bson:"created_at"`
  IsUsed      bool               `bson:"is_used"`
}

// JWT secret key from environment
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// Generate random verification code
func generateVerificationCode() string {
  rand.Seed(time.Now().UnixNano())
  randomNumber := rand.Intn(900000) + 100000
  return fmt.Sprintf("%06d", randomNumber)
}

// Generate JWT token
func generateJWTToken(user *user.User) (string, error) {
  expirationTime := time.Now().Add(24 * time.Hour)

  claims := &Claims{
    UserID:      user.ID.Hex(),
    Email:       user.Email,
    PhoneNumber: user.PhoneNumber,
    IsActive:    user.IsActive,
    RegisteredClaims: jwt.RegisteredClaims{
      ExpiresAt: jwt.NewNumericDate(expirationTime),
      IssuedAt:  jwt.NewNumericDate(time.Now()),
      Issuer:    "zapmanejo",
    },
  }

  token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
  return token.SignedString(jwtSecret)
}

// Validate JWT token
func validateJWTToken(tokenString string) (*Claims, error) {
  token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
    return jwtSecret, nil
  })

  if err != nil {
    return nil, err
  }

  if claims, ok := token.Claims.(*Claims); ok && token.Valid {
    return claims, nil
  }

  // @todo: Consider storing the token in the DB and validating it here.

  return nil, errors.New("invalid token")
}

// Store email verification code
func storeVerificationCode(userID primitive.ObjectID, email string) (string, error) {
  code := generateVerificationCode()

  verification := EmailVerification{
    UserID:    userID,
    Email:     email,
    Code:      code,
    ExpiresAt: time.Now().Add(15 * time.Minute), // 15 minute expiry
    CreatedAt: time.Now(),
    IsUsed:    false,
  }

  collection := db.GetCollection("email_verifications")
  _, err := collection.InsertOne(context.TODO(), verification)
  if err != nil {
    return "", err
  }

  return code, nil
}

// Verify email code
func verifyEmailCode(email, code string) (*user.User, error) {
  collection := db.GetCollection("email_verifications")

  var verification EmailVerification
  filter := bson.M{
    "email":      email,
    "code":       code,
    "is_used":    false,
    "expires_at": bson.M{"$gt": time.Now()},
  }

  err := collection.FindOne(context.TODO(), filter).Decode(&verification)
  if err != nil {
    return nil, errors.New("invalid or expired verification code")
  }

  // Mark verification as used
  update := bson.M{"$set": bson.M{"is_used": true}}
  collection.UpdateOne(context.TODO(), bson.M{"_id": verification.ID}, update)

  // Activate user account
  userCollection := db.GetCollection("users")
  userUpdate := bson.M{"$set": bson.M{"is_active": true}}
  userCollection.UpdateOne(context.TODO(), bson.M{"_id": verification.UserID}, userUpdate)

  // Return updated user
  var user user.User
  err = userCollection.FindOne(context.TODO(), bson.M{"_id": verification.UserID}).Decode(&user)
  return &user, err
}

// Link WhatsApp phone number to user account
func linkPhoneNumber(userID primitive.ObjectID, phoneNumber string) error {
  collection := db.GetCollection("users")
  update := bson.M{"$push": bson.M{"phone_numbers": phoneNumber}}
  _, err := collection.UpdateOne(context.TODO(), bson.M{"_id": userID}, update)
  return err
}

// Register new user handler
func HandleAuthRegister(w http.ResponseWriter, r *http.Request) {
  var req RegisterRequest
  if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    response := AuthResponse{Success: false, Message: err.Error()}
    json.NewEncoder(w).Encode(response)
    log.Printf("could not marshal request %v", err)
    return
  }

  // Validate input
  if req.Email == "" || req.Password == "" {
    response := AuthResponse{Success: false, Message: "Email and password are required"}
    json.NewEncoder(w).Encode(response)
    log.Printf("invalid input")
    return
  }

  emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
  if ! emailRegex.MatchString(req.Email) {
    response := AuthResponse{Success: false, Message: "Email is invalid"}
    json.NewEncoder(w).Encode(response)
    log.Printf("invalid email")
    return
  }

  collection := db.GetCollection("users")

  // Check if user already exists
  var existingUser user.User
  err := collection.FindOne(context.TODO(), bson.M{"email": req.Email}).Decode(&existingUser)
  if err == nil {
    response := AuthResponse{Success: false, Message: "User with this email already exists"}
    json.NewEncoder(w).Encode(response)
    log.Printf("user exists")
    return
  }

  hashedPassword, err := password.GetSalted(req.Password)
  if err != nil {
    response := AuthResponse{Success: false, Message: "Error processing password"}
    json.NewEncoder(w).Encode(response)
    log.Printf("could not hash password %v", err)
    return
  }

  // Create new user
  user := user.User{
    Name:         req.Name,
    Email:        req.Email,
    Password:     hashedPassword,
    CreatedAt:    time.Now(),
    UpdatedAt:    time.Now(),
    IsActive:     false, // Will be activated after email verification
    PhoneNumber:  req.PhoneNumber,
  }

  result, err := collection.InsertOne(context.TODO(), user)
  if err != nil {
    response := AuthResponse{Success: false, Message: "Error creating user account"}
    json.NewEncoder(w).Encode(response)
    log.Printf("could not insert user %v", err)
    return
  }

  // Get the inserted user ID
  userID := result.InsertedID.(primitive.ObjectID)
  user.ID = userID

  // Generate and store verification code
  verificationCode, err := storeVerificationCode(userID, req.Email)
  if err != nil {
    response := AuthResponse{Success: false, Message: "Error generating verification code"}
    json.NewEncoder(w).Encode(response)
    log.Printf("could not store verification code %v", err)
    return
  }

  // Special testing email domain.
  // @todo: make this an environment variable
  if strings.Contains(req.Email, "zapmanejo.test") {
    log.Printf("test email detected, skipping sending email registration")
    response := AuthResponse{
      Success: true,
      Message: "Registration successful. Please check your email for verification code.",
      User:    &user,
      VerificationCode: verificationCode,
    }
    json.NewEncoder(w).Encode(response)
    return
  }

  err = email.SendRegistrationEmail(req.Email, verificationCode)
  if err != nil {
    response := AuthResponse{Success: false, Message: "Error sending verification email"}
    json.NewEncoder(w).Encode(response)
    log.Printf("could not send registration email %v", err)
    return
  }

  response := AuthResponse{
    Success: true,
    Message: "Registration successful. Please check your email for verification code.",
    User:    &user,
  }
  json.NewEncoder(w).Encode(response)
}

// Login user handler
func HandleLogin(w http.ResponseWriter, r *http.Request) {
  if r.Method != "POST" {
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    return
  }

  var req LoginRequest
  if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    response := AuthResponse{Success: false, Message: "Invalid request format"}
    json.NewEncoder(w).Encode(response)
    log.Printf("could not marshal request: %v", err)
    return
  }

  hashedPassword, err := password.GetSalted(req.Password)
  if err != nil {
    response := AuthResponse{Success: false, Message: "Authentication failed"}
    json.NewEncoder(w).Encode(response)
    log.Printf("could not generate salted password: %v", err)
    return
  }

  // Find user with email and password
  collection := db.GetCollection("users")
  var user user.User
  filter := bson.M{
    "email":    req.Email,
    "password": hashedPassword,
  }

  err = collection.FindOne(context.TODO(), filter).Decode(&user)
  if err != nil {
    response := AuthResponse{Success: false, Message: "Invalid email or password"}
    json.NewEncoder(w).Encode(response)
    log.Printf("could not check for existing user: %v", err)
    return
  }

  // Check if account is active
  if !user.IsActive {
    response := AuthResponse{Success: false, Message: "Account not verified. Please check your email."}
    json.NewEncoder(w).Encode(response)
    return
  }

  // Generate JWT token
  token, err := generateJWTToken(&user)
  if err != nil {
    response := AuthResponse{Success: false, Message: "Error generating authentication token"}
    json.NewEncoder(w).Encode(response)
    log.Printf("could not generate auth token: %v", err)
    return
  }

  // Clear password from response
  user.Password = ""

  response := AuthResponse{
    Success: true,
    Message: "Login successful",
    Token:   token,
    User:    &user,
  }
  json.NewEncoder(w).Encode(response)
}

// Email verification handler
func HandleEmailVerification(w http.ResponseWriter, r *http.Request) {
  if r.Method != "POST" {
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    return
  }

  var req struct {
    Email string `json:"email"`
    Code  string `json:"code"`
  }

  if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    response := AuthResponse{Success: false, Message: "Invalid request format"}
    json.NewEncoder(w).Encode(response)
    return
  }

  user, err := verifyEmailCode(req.Email, req.Code)
  if err != nil {
    response := AuthResponse{Success: false, Message: err.Error()}
    json.NewEncoder(w).Encode(response)
    log.Printf("error verifying email code %v", err)
    return
  }

  // Generate JWT token for verified user
  token, err := generateJWTToken(user)
  if err != nil {
    response := AuthResponse{Success: false, Message: "Error generating authentication token"}
    json.NewEncoder(w).Encode(response)
    log.Printf("error generating token %v", err)
    return
  }

  // A new member is always a member of his own team
  collection := db.GetCollection("teams");
  data := map[string]string{
    "account": user.ID.Hex(),
    "name":  user.Name,
    "phone_number":  user.PhoneNumber, 
  }
  _, err = collection.InsertOne(context.TODO(), data)
  if err != nil {
    log.Printf("Error Inserting teams: %v", err)
    return 
  }

  // Update the users with account number
  users := db.GetCollection("users")
  filter := bson.M{"email":user.Email}
  update := bson.M{"$set": bson.M{"account": user.ID.Hex()}}
  users.UpdateOne(context.TODO(), filter, update)

  // Clear password from response
  user.Password = ""

  response := AuthResponse{
    Success: true,
    Message: "Email verified successfully",
    Token:   token,
    User:    user,
  }
  json.NewEncoder(w).Encode(response)
}

// Authentication middleware
func AuthMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    authHeader := r.Header.Get("Authorization")

    if authHeader == "" {
      // If we can't get it from the header, get it from the query string.
      // @todo:  I think this was for download links.  Is this needed?
      authHeader = r.URL.Query().Get("token")
    }

    if authHeader == "" {
      log.Printf("Missing Authorization Header")
      http.Error(w, "Authorization header required", http.StatusUnauthorized)
      return
    }

    tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
    claims, err := validateJWTToken(tokenString)
    if err != nil {
      log.Printf("Token is Invalid")
      http.Error(w, "Invalid authentication token", http.StatusUnauthorized)
      return
    }

    // Add user info to request context
    ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
    ctx = context.WithValue(ctx, "user_email", claims.Email)
    ctx = context.WithValue(ctx, "phone_number", claims.PhoneNumber)

    next.ServeHTTP(w, r.WithContext(ctx))
  })
}

func HandleGetUser(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
  userID := ctx.Value("user_id")
  if userID == nil {
    log.Printf("could not get userid from context")
    http.Error(w, "Authorization header required", http.StatusUnauthorized)
    return
  }
  log.Printf("Getting user from ID: %v", userID)

  user, err := user.Read(userID.(string))
  if err != nil {
    log.Printf("could not read userID from context")
    http.Error(w, "User Not Found", http.StatusNotFound)
    return
  }

  json, err := json.Marshal(user)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest) 
    fmt.Fprintf(w, "%v", err)
    return 
  }
  fmt.Fprint(w, string(json))
}


func HandleForgotPassword(w http.ResponseWriter, r *http.Request) {
  var req struct {
    Email string `json:"email"`
  }

  if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    response := AuthResponse{Success: false, Message: err.Error()}
    json.NewEncoder(w).Encode(response)
    log.Printf("could not marshal request %v", err)
    return
  }

  log.Printf("HandleForgotPassword");
}

func HandleChangePassword(w http.ResponseWriter, r *http.Request) {
  var req struct {
    CurrentPassword string `json:"current_password"`
    NewPassword string `json:"new_password"`
  }

  if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    response := AuthResponse{Success: false, Message: err.Error()}
    json.NewEncoder(w).Encode(response)
    log.Printf("could not marshal request %v", err)
    return
  }

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

  log.Printf("Getting user from ID: %v", userID)
  log.Printf("HandleChangePassword");
  log.Printf("Current: %s", req.CurrentPassword);
  log.Printf("New: %s", req.NewPassword);

  reqPass, err := password.GetSalted(req.CurrentPassword)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest) 
    fmt.Fprintf(w, "%v", err)
    return
  }

  if user.Password != reqPass {
    log.Printf("Passed current password does not match currend db password")
    w.WriteHeader(http.StatusBadRequest) 
    fmt.Fprintf(w, "%v", "Current password does not match")
    return
  }

  newPass, err := password.GetSalted(req.NewPassword)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest) 
    fmt.Fprintf(w, "%v", err)
    return
  }

  err = user.Update("password", newPass)
  if err != nil {
    http.Error(w, "Error Updating Data", http.StatusBadRequest)
    log.Printf("Error Updating Data: %v", err)
    return 
  }

}
