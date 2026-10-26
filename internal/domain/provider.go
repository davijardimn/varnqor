package domain

import "time"

type ProviderCapabilities struct {
	Usage  bool `json:"usage"`
	Tools  bool `json:"tools"`
	Limits bool `json:"limits"`
}

type ProviderStatus struct {
	ID               string               `json:"id"`
	Name             string               `json:"name"`
	State            string               `json:"state"`
	Detected         bool                 `json:"detected"`
	BinaryDetected   bool                 `json:"binaryDetected"`
	DataDetected     bool                 `json:"dataDetected"`
	HomePath         string               `json:"homePath,omitempty"`
	DataPath         string               `json:"dataPath,omitempty"`
	ObservedSessions int                  `json:"observedSessions"`
	ObservedProjects int                  `json:"observedProjects"`
	IndexedCalls     int                  `json:"indexedCalls"`
	TokensToday      int64                `json:"tokensToday"`
	FilesIndexed     int                  `json:"filesIndexed"`
	ParseWarnings    int                  `json:"parseWarnings"`
	UnsupportedFiles int                  `json:"unsupportedFiles"`
	LastActivity     time.Time            `json:"lastActivity,omitempty"`
	Capabilities     ProviderCapabilities `json:"capabilities"`
	Note             string               `json:"note,omitempty"`
	Error            string               `json:"error,omitempty"`
}
