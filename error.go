package main

/* Descriptive error struct to inform the user for all the error issues */
type errorHandler struct {
	Error   error
	Message string
	Code    int
}
