package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

func ledgerMetrics(ledger []model.LedgerRecord, emitted int64) InstrumentMetrics {
	m := InstrumentMetrics{LedgerComplete: true, Emitted: emitted}
	ids, sequences := map[uint64]bool{}, map[int64]bool{}
	for _, row := range ledger {
		if row.DeliveryID == 0 || ids[row.DeliveryID] {
			m.LedgerComplete = false
		}
		ids[row.DeliveryID], sequences[row.Seq] = true, true
		countDelivery(&m, row)
	}
	m.LedgerComplete = m.LedgerComplete && containsEveryEmission(sequences, emitted)
	return m
}

func countDelivery(m *InstrumentMetrics, row model.LedgerRecord) {
	switch row.DeliveryReason {
	case model.DeliveryDroppedByPerturb:
		m.Dropped++
	case model.DeliveryDuplicated:
		m.Duplicated++
	case model.DeliveryMangled:
		m.Mangled++
	case model.DeliveryDelayed:
		m.Delayed++
	}
	if row.Delivered {
		m.Delivered++
	}
}

func containsEveryEmission(sequences map[int64]bool, emitted int64) bool {
	for seq := int64(0); seq < emitted; seq++ {
		if !sequences[seq] {
			return false
		}
	}
	return true
}
