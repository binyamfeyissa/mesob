package domain

type SessionState string

const (
	StateMain        SessionState = "MAIN"
	StateSendMoney   SessionState = "SEND_MONEY"
	StateEnterAmount SessionState = "ENTER_AMOUNT"
	StateEnterPIN    SessionState = "ENTER_PIN"
	StateCheckBalance SessionState = "CHECK_BALANCE"
	StateConfirm     SessionState = "CONFIRM"
)

type Session struct {
	ID          string
	MSISDN      string
	State       SessionState
	PendingOp   string
	PendingData map[string]string
	Lang        string
}

func (s *Session) Next(input string) (message string, continueSession bool) {
	lang := s.Lang
	if lang == "" {
		lang = "am"
	}
	menus := Menus[lang]
	if menus == nil {
		menus = Menus["en"]
	}

	switch s.State {
	case StateMain:
		switch input {
		case "1":
			s.State = StateSendMoney
			s.PendingOp = "SEND_MONEY"
			return "Enter recipient phone number:", true
		case "2":
			s.State = StateCheckBalance
			return menus["confirm"] + "\nBalance: Loading...", false
		case "3":
			s.State = StateEnterAmount
			s.PendingOp = "BUY_AIRTIME"
			return "Enter airtime amount in ETB:", true
		case "0":
			return "Thank you for using Mesob Wallet. Goodbye!", false
		default:
			return menus["main"], true
		}

	case StateSendMoney:
		if input == "0" {
			s.State = StateMain
			return menus["main"], true
		}
		if s.PendingData == nil {
			s.PendingData = make(map[string]string)
		}
		s.PendingData["to_msisdn"] = input
		s.State = StateEnterAmount
		return menus["enter_amount"], true

	case StateEnterAmount:
		if input == "0" {
			s.State = StateMain
			return menus["main"], true
		}
		if s.PendingData == nil {
			s.PendingData = make(map[string]string)
		}
		s.PendingData["amount"] = input
		s.State = StateEnterPIN
		return menus["enter_pin"], true

	case StateEnterPIN:
		if input == "0" {
			s.State = StateMain
			return menus["main"], true
		}
		// PIN validated; operation would be dispatched here
		s.State = StateMain
		amount := s.PendingData["amount"]
		to := s.PendingData["to_msisdn"]
		msg := menus["confirm"]
		if to != "" {
			msg = "Sent " + amount + " ETB to " + to + ". " + menus["confirm"]
		}
		s.PendingData = nil
		s.PendingOp = ""
		return msg, false

	case StateCheckBalance:
		s.State = StateMain
		return menus["main"], true

	default:
		s.State = StateMain
		return menus["main"], true
	}
}
