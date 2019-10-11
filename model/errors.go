package model

type EntryNotFoundError struct {
	Message string
}

func (err EntryNotFoundError) Error() string {
	if len(err.Message) == 0 {
		return "Entry not found."
	} else {
		return err.Message
	}
}
