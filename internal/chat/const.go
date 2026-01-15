package chat

// Breed of Animal
const ANGUS = "angus"
const NELORE = "nelore"
const BRANGUS = "brangus"
const STA_ZELIA = "sta.zelia"
const CRUZADA = "cruzada"
const CRUZADO = "cruzado"
const MURRAH = "murrah"
const MEDITERRANEO = "mediterrâneo"
const JAFARABADI = "jafarabadi"
const CARABAO = "carabao"

var BREEDS = []string{
	ANGUS, NELORE, BRANGUS, STA_ZELIA,
	CRUZADA, CRUZADO, MURRAH, MEDITERRANEO,
	JAFARABADI, CARABAO,
}

// Sex of Animal
const MALE = "m"
const FEMALE = "f"
var SEXES = []string{MALE, FEMALE}

// Dead Types
const MORREU     = "morreu"       // he died
const MORTO      = "morto"        // dead
const NASCEU     = "nasceu morto" // stillborn (born dead)
const ABORTO     = "aborto"       // aborted
const NATIMORTO  = "natimorto"    // stillborn
const NATIMORTOS = "natimortos"   // stillbirths
var DEATHS = []string{MORREU, MORTO, NASCEU, ABORTO, NATIMORTO, NATIMORTOS}

// Pure Breed Specifier
const PURE_BREED = "fft" 
