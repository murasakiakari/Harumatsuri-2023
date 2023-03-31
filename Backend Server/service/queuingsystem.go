package service

import (
	"backendserver/utility"
	"encoding/json"
	"fmt"
	"regexp"
	"sync"
	"time"
)

var (
	queuingSystem = NewQueuingSystem()
	pattern       = regexp.MustCompile(`^R?\d+$`)
)

type Entry int

const (
	_ Entry = iota
	FIRST
	SECOND
)

type QueuingNumber struct {
	entry  Entry
	number int
}

func (data *QueuingNumber) String() string {
	switch data.entry {
	case FIRST:
		return fmt.Sprintf("%v", data.number)
	case SECOND:
		return fmt.Sprintf("R%v", data.number)
	}
	return "invalid"
}

type EntryPair struct {
	First  int `json:"first"`
	Second int `json:"second"`
}

type QueuingInformation struct {
	Current         EntryPair `json:"current"`
	Issued          EntryPair `json:"issued"`
	WaitingTime     float64   `json:"waitingTime"`
	Offset          float64   `json:"offset"`
	TotalAllowEntry int       `json:"totalAllowEntry"`
}

type AdmitRecord struct {
	actionTime time.Time
	AllowEntry int
}

type QueuingSystem struct {
	started            bool
	queue              []*QueuingNumber
	queueOrder         map[string]int
	queuingInformation *QueuingInformation
	queuingHistory     []*AdmitRecord
	lock               sync.RWMutex
}

func NewQueuingSystem() *QueuingSystem {
	return &QueuingSystem{
		started:            false,
		queue:              make([]*QueuingNumber, 0, 1024),
		queueOrder:         make(map[string]int),
		queuingInformation: &QueuingInformation{Current: EntryPair{First: -1, Second: -1}, Issued: EntryPair{First: utility.Config.Ticket.Offset}, WaitingTime: -1},
		queuingHistory:     make([]*AdmitRecord, 0, 32),
	}
}

func (s *QueuingSystem) String() string {
	return fmt.Sprintf("queue: %v\nqueueOrder: %v\nqueuingInformation: %v", s.queue, s.queueOrder, s.queuingInformation)
}

func (s *QueuingSystem) encodeMessage() ([]byte, error) {
	message, err := json.Marshal(s.queuingInformation)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func (s *QueuingSystem) StartQueuing(peopleAdmitted int) ([]byte, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.started {
		return s.encodeMessage()
	}

	s.started = true

	upperBound := s.queuingInformation.TotalAllowEntry + peopleAdmitted
	if len(s.queue) < upperBound {
		upperBound = len(s.queue)
	}

	s.queuingInformation.Current.First = 0
	s.queuingInformation.Current.Second = 0

	for i := s.queuingInformation.TotalAllowEntry; i < upperBound; i++ {
		queuingData := s.queue[i]
		switch queuingData.entry {
		case FIRST:
			s.queuingInformation.Current.First = queuingData.number
		case SECOND:
			s.queuingInformation.Current.Second = queuingData.number
		}
	}

	s.queuingInformation.WaitingTime = 0.6
	s.queuingInformation.TotalAllowEntry += peopleAdmitted
	s.queuingHistory = append(s.queuingHistory, &AdmitRecord{actionTime: time.Now(), AllowEntry: peopleAdmitted})
	return s.encodeMessage()
}

func (s *QueuingSystem) AllowEntry(peopleAdmitted int) ([]byte, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if !s.started {
		return s.encodeMessage()
	}

	upperBound := s.queuingInformation.TotalAllowEntry + peopleAdmitted
	if len(s.queue) < upperBound {
		upperBound = len(s.queue)
	}

	for i := s.queuingInformation.TotalAllowEntry; i < upperBound; i++ {
		queuingData := s.queue[i]
		switch queuingData.entry {
		case FIRST:
			s.queuingInformation.Current.First = queuingData.number
		case SECOND:
			s.queuingInformation.Current.Second = queuingData.number
		}
	}

	previousActionIndex := len(s.queuingHistory) - 1
	previousAction := s.queuingHistory[previousActionIndex]
	s.queuingInformation.WaitingTime = float64(time.Since(previousAction.actionTime)/time.Minute) / float64(peopleAdmitted)
	if s.queuingInformation.WaitingTime < 0.6 {
		s.queuingInformation.WaitingTime = 0.6
	}
	s.queuingInformation.TotalAllowEntry += peopleAdmitted
	s.queuingHistory = append(s.queuingHistory, &AdmitRecord{actionTime: time.Now(), AllowEntry: peopleAdmitted})
	return s.encodeMessage()
}

func (s *QueuingSystem) SetOffset(offset float64) ([]byte, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if !s.started {
		return s.encodeMessage()
	}

	s.queuingInformation.Offset = offset
	return s.encodeMessage()
}

func (s *QueuingSystem) UpdateFirstEntry(peopleEntries int) ([]byte, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if peopleEntries < 1 {
		return s.encodeMessage()
	}

	for i := 0; i < peopleEntries; i, s.queuingInformation.Issued.First = i+1, s.queuingInformation.Issued.First+1 {
		queuingNumber := &QueuingNumber{entry: FIRST, number: s.queuingInformation.Issued.First + 1}
		s.queue = append(s.queue, queuingNumber)
		s.queueOrder[queuingNumber.String()] = len(s.queue)
	}

	if s.queuingInformation.TotalAllowEntry > 0 && s.queuingInformation.TotalAllowEntry >= len(s.queue) {
		index := len(s.queue) - 1
		s.queuingInformation.Current.First = s.queue[index].number
	}

	return s.encodeMessage()
}

func (s *QueuingSystem) UpdateSecondEntry(peopleEntries int) ([]byte, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if peopleEntries < 1 {
		return s.encodeMessage()
	}

	for i := 0; i < peopleEntries; i, s.queuingInformation.Issued.Second = i+1, s.queuingInformation.Issued.Second+1 {
		queuingNumber := &QueuingNumber{entry: SECOND, number: s.queuingInformation.Issued.Second + 1}
		s.queue = append(s.queue, queuingNumber)
		s.queueOrder[queuingNumber.String()] = len(s.queue)
	}

	if s.queuingInformation.TotalAllowEntry > 0 && s.queuingInformation.TotalAllowEntry >= len(s.queue) {
		index := len(s.queue) - 1
		s.queuingInformation.Current.Second = s.queue[index].number
	}

	return s.encodeMessage()
}

func (s *QueuingSystem) GetOrder(number string) int {
	if !pattern.Match([]byte(number)) {
		return -1
	}

	s.lock.RLock()
	defer s.lock.RUnlock()
	order, ok := s.queueOrder[number]
	if !ok {
		return -1
	}
	return order
}

func (s *QueuingSystem) GetMessage() ([]byte, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.encodeMessage()
}
