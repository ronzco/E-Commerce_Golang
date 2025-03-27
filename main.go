package main
import (
"encoding/json"
"fmt"
"io"
"log"
"net/http"
"os"
"path/filepath"
"project/internal/delivery"
"project/internal/models"
"project/internal/repository"
"project/internal/usecase"
"time"
"github.com/gorilla/mux"
"github.com/joho/godotenv"
"gorm.io/driver/mysql"
"gorm.io/gorm"
)
var jwtKey = []byte(getEnv("JWT_SECRET", "IniKunciRahasiaSuperAman123!@#"))
func getEnv(key, fallback string) string {
_ = godotenv.Load()
if value, exists := os.LookupEnv(key); exists {
return value
}
return fallback
}
var db *gorm.DB
func InitDB() {
dsn := "root:@tcp(localhost:3306)/ecommerce?charset=utf8mb4&parseTime=True&loc=Local"
var err error
db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
if err != nil {
log.Fatal("Failed to connect to database", err)
}
db.AutoMigrate(&models.User{}, &models.Store{}, &models.Address{}, &models.Category{},
&models.Product{}, &models.Transaction{}, &models.LogProduct{})
}
func GetUserAddressesHandler(w http.ResponseWriter, r *http.Request) {
userID, ok := r.Context().Value("userID").(uint)
if !ok {
http.Error(w, "Unauthorized", http.StatusUnauthorized)
return
}
var addresses []models.Address
db.Where("user_id = ?", userID).Find(&addresses)
json.NewEncoder(w).Encode(addresses)
}
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
w.Write([]byte("Update user endpoint"))
}
func CreateAddressHandler(w http.ResponseWriter, r *http.Request) {
var address models.Address
if err := json.NewDecoder(r.Body).Decode(&address); err != nil {
http.Error(w, "Invalid request", http.StatusBadRequest)
return
}
if address.UserID == 0 {
http.Error(w, "User ID is required", http.StatusBadRequest)
return
}
if err := db.Create(&address).Error; err != nil {
http.Error(w, "Failed to create address", http.StatusInternalServerError)
return
}
w.WriteHeader(http.StatusCreated)
json.NewEncoder(w).Encode(address)
}
// CRUD Handlers for Category (Admin Only)
func CreateCategoryHandler(w http.ResponseWriter, r *http.Request) {
var category models.Category
if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
http.Error(w, "Invalid request", http.StatusBadRequest)
return
}
db.Create(&category)
w.WriteHeader(http.StatusCreated)
json.NewEncoder(w).Encode(category)
}
func UploadProductImageHandler(w http.ResponseWriter, r *http.Request) {
r.ParseMultipartForm(10 << 20) // 10 MB limit
file, handler, err := r.FormFile("image")
if err != nil {
http.Error(w, "Error retrieving the file", http.StatusBadRequest)
return
}
defer file.Close()
dir := "uploads/"
if _, err := os.Stat(dir); os.IsNotExist(err) {
os.Mkdir(dir, os.ModePerm)
}
fileName := fmt.Sprintf("%d%s", time.Now().Unix(), filepath.Ext(handler.Filename))
filePath := filepath.Join(dir, fileName)
outFile, err := os.Create(filePath)
if err != nil {
http.Error(w, "Error saving file", http.StatusInternalServerError)
return
}
defer outFile.Close()
io.Copy(outFile, file)
json.NewEncoder(w).Encode(map[string]string{"image_url": filePath})
}
func main() {
InitDB()
// Set up repositories
userRepo := repository.NewUserRepository(db)
userUsecase := usecase.NewUserUsecase(userRepo, jwtKey)
userHandler := delivery.NewUserHandler(userUsecase)
productRepo := repository.NewProductRepository(db)
transactionRepo := repository.NewTransactionRepository(db)
// Set up repository dan usecase log product
logProductRepo := repository.NewLogProductRepository(db)
logProductUsecase := usecase.NewLogProductUsecase(logProductRepo)
logProductHandler := delivery.NewLogProductHandler(logProductUsecase)
// Set up use cases
productUsecase := usecase.NewProductUsecase(productRepo)
transactionUsecase := usecase.NewTransactionUsecase(productRepo, transactionRepo)
// Set up handlers
productHandler := delivery.NewProductHandler(productUsecase)
transactionHandler := delivery.NewTransactionHandler(transactionUsecase)
// Setup router
router := mux.NewRouter()
// PROVINCE DATA ROUTE (External API)
router.HandleFunc("/provinces", delivery.FetchProvincesHandler).Methods("GET")
// USER ROUTES
router.HandleFunc("/register", userHandler.RegisterHandler).Methods("POST")
router.HandleFunc("/login", userHandler.LoginHandler).Methods("POST")
router.Handle("/user",
delivery.AuthMiddleware(userUsecase)(http.HandlerFunc(UpdateUserHandler))).Methods("PUT")
router.Handle("",
delivery.AuthMiddleware(userUsecase)(http.HandlerFunc(UpdateUserHandler))).Methods("PUT")
// UPLOAD ROUTES
router.Handle("/upload",
delivery.AuthMiddleware(userUsecase)(http.HandlerFunc(UploadProductImageHandler))).Methods("P
OST")
router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/",
http.FileServer(http.Dir("uploads/"))))
// ADDRESS ROUTES
addressRouter := router.PathPrefix("/addresses").Subrouter()
addressRouter.Handle("",
delivery.AuthMiddleware(userUsecase)(http.HandlerFunc(CreateAddressHandler))).Methods("POST")
addressRouter.Handle("",
delivery.AuthMiddleware(userUsecase)(http.HandlerFunc(GetUserAddressesHandler))).Methods("GET
")
// PRODUCT ROUTES (semua user bisa)
productRouter := router.PathPrefix("/products").Subrouter()
productRouter.Handle("",
http.HandlerFunc(productHandler.GetProductsHandler)).Methods("GET")
productRouter.Handle("",
delivery.AuthMiddleware(userUsecase)(http.HandlerFunc(productHandler.CreateProductHandler))).
Methods("POST")
// TRANSACTION ROUTES
transactionRouter := router.PathPrefix("/transactions").Subrouter()
transactionRouter.Handle("",
delivery.AuthMiddleware(userUsecase)(http.HandlerFunc(transactionHandler.CreateTransactionHan
dler))).Methods("POST")
// CATEGORY ROUTES (Admin Only)
categoryRouter := router.PathPrefix("/categories").Subrouter()
categoryRouter.Handle("", delivery.AuthMiddleware(userUsecase)(
delivery.AdminOnlyMiddleware(userUsecase)(
http.HandlerFunc(CreateCategoryHandler),
),
)).Methods("POST")
// LOG PRODUCT ROUTES
logProductRouter := router.PathPrefix("/log-products").Subrouter()
logProductRouter.Handle("",
delivery.AuthMiddleware(userUsecase)(http.HandlerFunc(logProductHandler.GetLogProductsHandler
))).Methods("GET")
// ADMIN ONLY PRODUCT ROUTE
router.Handle("/admin/products",
delivery.AdminOnlyMiddleware(userUsecase)(http.HandlerFunc(productHandler.CreateProductHandle
r))).Methods("POST")
fmt.Println("🚀 Server running on port 8080")
log.Fatal(http.ListenAndServe(":8080", router))
}
