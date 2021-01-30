package app_msgs

import "fmt"

// Errors
const (
	msCredentials   = "error: microsoft credentials environment variables must be set"
	authCodeMissing = "error: authorization code was not provided"
	notAllTasksParsed = "error: not all tasks could be parsed. Parsed %v tasks of %v\n"
	notAllTransactionsStored = "error: not all transactions could be stored. Saved %v transactions of %v\n"
)

const (
	allTransactionsStored = "all %v transactions were stored successfully!\n"
	allTasksDeleted = "deleted %v tasks!\n"
)

func MsCredentials() string {
	return msCredentials
}

func AuthCodeMissing() string {
	return authCodeMissing
}

func NotAllTasksParsed(parsed int, total int) string {
	return fmt.Sprintf(notAllTasksParsed, parsed, total)
}

func NotAllTransactionsStored(saved int, total int) string {
	return fmt.Sprintf(notAllTransactionsStored, saved, total)
}

func AllTransactionsStored(count int) string {
	return fmt.Sprintf(allTransactionsStored, count)
}

func AllTasksDeleted(count int) string {
	return fmt.Sprintf(allTasksDeleted, count)
}
