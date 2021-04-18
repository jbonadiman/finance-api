package app_msgs

import "fmt"

// Errors
const (
	msCredentials            = "error: microsoft credentials environment variables must be set"
	authCodeMissing          = "error: authorization code was not provided"
	notAllTasksParsed        = "error: not all tasks could be parsed. Parsed %v tasks of %v\n"
	notAllTransactionsStored = "error: not all transactions could be stored. Saved %v transactions of %v\n"
	errorCompletingTasks     = "an error occurred while marking tasks as completed: %v\n"
	invalidAuthState         = "invalid auth state: %v\n"
	errorAuthenticating      = "an error occurred during authentication: %v\n"
	redisConnectionError     = "an error occurred while connecting to Redis: %v\n"
	notAuthenticated         = "there is no token information saved. Please, authenticate"
)

// Successes
const (
	allTransactionsStored = "all %v transactions were stored successfully!\n"
	allTasksCompleted     = "marked %v tasks as completed!\n"
)

func MsCredentials() string {
	return msCredentials
}

func NotAuthenticated() string {
	return notAuthenticated
}

func AuthCodeMissing() string {
	return authCodeMissing
}

func InvalidAuthState(state string) string {
	return fmt.Sprintf(invalidAuthState, state)
}

func RedisConnectionError(error string) string {
	return fmt.Sprintf(redisConnectionError, error)
}

func ErrorAuthenticating(error string) string {
	return fmt.Sprintf(errorAuthenticating, error)
}

func NotAllTasksParsed(parsed int, total int) string {
	return fmt.Sprintf(notAllTasksParsed, parsed, total)
}

func NotAllTransactionsStored(saved int, total int) string {
	return fmt.Sprintf(notAllTransactionsStored, saved, total)
}

func ErrorCompletingTasks(errText string) string {
	return fmt.Sprintf(errorCompletingTasks, errText)
}

func AllTransactionsStored(count int) string {
	return fmt.Sprintf(allTransactionsStored, count)
}

func AllTasksCompleted(count int) string {
	return fmt.Sprintf(allTasksCompleted, count)
}
