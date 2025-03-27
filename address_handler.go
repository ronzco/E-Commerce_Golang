package delivery
import (
"encoding/json"
"net/http"
"project/internal/models"
"project/internal/usecase"
"strconv"
"github.com/gorilla/mux"
)
type AddressHandler struct {
usecase usecase.AddressUsecase
}
func NewAddressHandler(usecase usecase.AddressUsecase) *AddressHandler {
return &AddressHandler{usecase: usecase}
}
// Buat alamat baru
func (h *AddressHandler) CreateAddressHandler(w http.ResponseWriter, r *http.Request) {
userID, ok := r.Context().Value(UserIDKey).(uint)
if !ok {
http.Error(w, "Unauthorized", http.StatusUnauthorized)
return
}
var address models.Address
if err := json.NewDecoder(r.Body).Decode(&address); err != nil {
http.Error(w, "Invalid request", http.StatusBadRequest)
return
}
address.UserID = userID
if err := h.usecase.CreateAddress(&address); err != nil {
http.Error(w, "Failed to create address", http.StatusInternalServerError)
return
}
w.WriteHeader(http.StatusCreated)
json.NewEncoder(w).Encode(address)
}
// Lihat daftar alamat user
func (h *AddressHandler) GetUserAddressesHandler(w http.ResponseWriter, r *http.Request) {
userID, ok := r.Context().Value(UserIDKey).(uint)
if !ok {
http.Error(w, "Unauthorized", http.StatusUnauthorized)
return
}
addresses, err := h.usecase.GetAddressesByUserID(userID)
if err != nil {
http.Error(w, "Failed to fetch addresses", http.StatusInternalServerError)
return
}
json.NewEncoder(w).Encode(addresses)
}
// Update alamat user
func (h *AddressHandler) UpdateAddressHandler(w http.ResponseWriter, r *http.Request) {
userID, ok := r.Context().Value(UserIDKey).(uint)
if !ok {
http.Error(w, "Unauthorized", http.StatusUnauthorized)
return
}
vars := mux.Vars(r)
addressID, err := strconv.Atoi(vars["id"])
if err != nil {
http.Error(w, "Invalid address ID", http.StatusBadRequest)
return
}
var address models.Address
if err := json.NewDecoder(r.Body).Decode(&address); err != nil {
http.Error(w, "Invalid request", http.StatusBadRequest)
return
}
address.ID = uint(addressID)
address.UserID = userID
if err := h.usecase.UpdateAddress(&address); err != nil {
http.Error(w, "Failed to update address", http.StatusInternalServerError)
return
}
json.NewEncoder(w).Encode(address)
}
// Hapus alamat user
func (h *AddressHandler) DeleteAddressHandler(w http.ResponseWriter, r *http.Request) {
userID, ok := r.Context().Value(UserIDKey).(uint)
if !ok {
http.Error(w, "Unauthorized", http.StatusUnauthorized)
return
}
vars := mux.Vars(r)
addressID, err := strconv.Atoi(vars["id"])
if err != nil {
http.Error(w, "Invalid address ID", http.StatusBadRequest)
return
}
if err := h.usecase.DeleteAddress(uint(addressID), userID); err != nil {
http.Error(w, "Failed to delete address", http.StatusInternalServerError)
return
}
w.WriteHeader(http.StatusNoContent)
}
