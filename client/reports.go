package client

// TODO: WHERE TO PICK UP: We need to create the ShortID file for the client
// within the test suite, and we need to authorize the hardware with the GCA
// server in the SetupTestEnvironment function.
//
// Then we need to keep making progress on this file. Currently we submit
// reports but not all of the functions are complete, then we'll need to test
// that, then we'll need to write and test the syncing function. Finally after
// that we'll be able to move on to writing failover code, migration code, and
// server reliability code.

// reports.go contains all of the code for sending reports to the server.
//
// TODO: We will have to create a simlink between the client directory's
// 'energy_data.csv' and the '/opt/halki/energy_data.csv' file.
//
// TODO: Need to confirm with the monitoring guys that every 5 minute period
// will get exactly 1 record in the CSV, and that every second is accounted for
// in some report or another.
//
// TODO: We're going to need to handle the case where two reports from the
// monitoring equipment end up in the same timeslot. Probably the correct
// solution is to roll the second reading into the next unread timeslot. We'll
// just have to have some way to tell that some report got squished into the
// wrong timeslot.

import (
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/glowlabs-org/gca-backend/glow"
)

// Represents one row of data from the energy file.
type EnergyRecord struct {
	Timeslot uint32
	Energy   uint64
}

// Create an energy report from the provided energy record.
func (c *Client) sendReport(er EnergyRecord) {
	eqr := glow.EquipmentReport{
		ShortID:     c.shortID,
		Timeslot:    er.Timeslot,
		PowerOutput: er.Energy,
	}
	sb := eqr.SigningBytes()
	eqr.Signature = glow.Sign(sb, c.privkey)
	data := eqr.Serialize()

	c.serverMu.Lock()
	gcas := c.gcaServers[c.primaryServer]
	location := fmt.Sprintf("%v:%v", gcas.Location, gcas.UdpPort)
	c.serverMu.Unlock()

	glow.SendUDPReport(data, location)
}

// readEnergyFile will read the data from the energy file and return an array
// that contains all of the values.
func (c *Client) readEnergyFile() ([]EnergyRecord, error) {
	// Open the CSV file
	filePath := path.Join(c.baseDir, EnergyFile)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open monitoring file: %v", err)
	}
	defer file.Close()

	// Iterate over the CSV records
	reader := csv.NewReader(file)
	var records []EnergyRecord
	for {
		record, err := reader.Read()
		if err != nil {
			break // Stop at EOF or on error
		}

		timestamp, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			continue // Skip records with invalid timestamps
		}
		timeslot, err := glow.UnixToTimeslot(timestamp)
		if err != nil {
			continue
		}
		energy, err := strconv.ParseUint(record[1], 10, 32)
		if err != nil {
			energy = 0
		}

		// Round the energy down so that we never over-estiamte the amount of
		// power that has been produced.
		energy = energy - 1

		// 0, 1, and 2 are reserved sentinel values, so we just skip this
		// reading if we are in that range.
		if energy < 3 {
			continue
		}

		// Append the data to the records slice
		records = append(records, EnergyRecord{
			Timeslot: timeslot,
			Energy:   uint64(energy - 1),
		})
	}

	return records, nil
}

// threadedSendReports will wake up every minute, check whether there's a new
// report available, and if so it'll send a report for the corresponding
// timeslot.
func (c *Client) threadedSendReports() {
	// Right at startup, we save all of the existing records.
	latestRecord := uint32(0)
	records, err := c.readEnergyFile()
	// We'll no-op if there's an error.
	if err == nil {
		for _, record := range records {
			err := c.saveReading(record.Timeslot, uint32(record.Energy))
			if err != nil {
				continue
			}
			if record.Timeslot > latestRecord {
				latestRecord = record.Timeslot
			}
		}
	}

	// Infinite loop to send reports.
	ticks := 0
	for {
		// Quit if the closeChan was closed.
		select {
		case <-c.closeChan:
			return
		default:
		}

		// Read the energy file. No-op if there's an error. Can't
		// continue because we still want to sleep.
		records, err := c.readEnergyFile()
		if err == nil {
			for _, record := range records {
				// We try saving the reading first, which can
				// produce an error. The main error that we are
				// looking for is a double report error, which
				// means the same timeslot has multiple
				// different energy readings. That's a problem
				// that will cause the timeslot to get banned,
				// so we don't send the report if that happens.
				err := c.saveReading(record.Timeslot, uint32(record.Energy))
				if err != nil {
					continue
				}
				if record.Timeslot > latestRecord {
					c.sendReport(record)
				}
			}
			// The above loop doesn't update the latestRecord
			// because if there are multiple new records we want to
			// send all of them.
			for _, record := range records {
				if record.Timeslot > latestRecord {
					latestRecord = record.Timeslot
				}
			}
		}

		// Sleep for a minute before checking again.
		//
		// TODO: Need to turn this into a soft sleep, and need to
		// change the duration for testing.
		time.Sleep(time.Minute)

		ticks++
		if ticks >= 300 {
			ticks = 0
			// TODO: go syncWithServer()
		}
	}
}