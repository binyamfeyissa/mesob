package domain

// Menu strings per language. No hard-coded strings in handlers.
var Menus = map[string]map[string]string{
	"am": {
		"main":         "እንኳን ደህና መጡ ወደ Mesob Wallet\n1. ገንዘብ ላክ\n2. ቀሪ ሂሳብ\n3. አየር ጊዜ ግዛ\n0. ውጣ",
		"enter_amount": "ምን ያህል ብር ማዘዝ ይፈልጋሉ?",
		"enter_pin":    "ፒን ቁጥርዎን ያስገቡ:",
		"confirm":      "ሂደቱ ተጠናቋል",
		"error":        "ስህተት ተፈጠረ። እባክዎ ዳግም ይሞክሩ",
	},
	"en": {
		"main":         "Welcome to Mesob Wallet\n1. Send Money\n2. Check Balance\n3. Buy Airtime\n0. Exit",
		"enter_amount": "Enter amount in ETB:",
		"enter_pin":    "Enter your PIN:",
		"confirm":      "Transaction complete",
		"error":        "An error occurred. Please try again",
	},
	"om": {
		"main": "Mesob Wallet bira baga nagaan dhuftan\n1. Maatii ergi\n2. Ilaalii\n3. Yeroo bilbilaa bitadhu\n0. Ba'i",
	},
	"ti": {
		"main": "እንኳዕ ናብ Mesob Wallet ብደሓን መጻኹም\n1. ገንዘብ ስደድ\n2. ቀሪ ሒሳብ\n3. ናይ ኤር ግዜ ዕድጊ\n0. ውጻእ",
	},
}
