package domain

const (
	StatusNew               = "NEW"
	StatusUnderReview       = "UNDER_REVIEW"
	StatusAwaitingDocuments = "AWAITING_DOCUMENTS"
	StatusQuoteReady        = "QUOTE_READY"
	StatusQuoteAccepted     = "QUOTE_ACCEPTED"
	StatusOrderProcessing   = "ORDER_PROCESSING"
	StatusInTransit         = "IN_TRANSIT"
	StatusArrived           = "ARRIVED"
	StatusCleared           = "CLEARED"
	StatusReadyForPickup    = "READY_FOR_PICKUP"
	StatusCompleted         = "COMPLETED"
	StatusCancelled         = "CANCELLED"
)

var AllowedStatuses = map[string]struct{}{
	StatusNew: {}, StatusUnderReview: {}, StatusAwaitingDocuments: {}, StatusQuoteReady: {}, StatusQuoteAccepted: {},
	StatusOrderProcessing: {}, StatusInTransit: {}, StatusArrived: {}, StatusCleared: {}, StatusReadyForPickup: {},
	StatusCompleted: {}, StatusCancelled: {},
}
