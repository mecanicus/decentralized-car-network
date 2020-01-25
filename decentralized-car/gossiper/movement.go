package gossiper

import (
	// 	"errors"
	// 	"fmt"
	// 	"log"
	// 	"math/rand"
	// 	"net"
	// 	"strconv"
	// 	"sync"
	// 	"time"

	// 	"github.com/dedis/protobuf"
	// 	"github.com/tormey97/decentralized-car-network/decentralized-car/messaging"

	"math/rand"
	"time"

	"github.com/tormey97/decentralized-car-network/utils"
)

func (peerster *Peerster) MoveCarPosition() {

	go func() {
		for {
			// If it is a police car stopped don't do anything
			if peerster.PathCar != nil {
				time.Sleep(time.Duration(peerster.BroadcastTimer) * time.Second)
				areaChange := peerster.changeOfArea()
				//There is a change in the area zone, so different procedure
				if areaChange {
					peerster.sendAreaChangeMessage(peerster.PathCar[1])
					peerster.startAreaChangeSession()
				} else {
					//This function will advance the car to the next position if possible, checking there are not other cars
					peerster.positionAdvancer()
				}
			}
		}
	}()
}

func (peerster *Peerster) startAreaChangeSession() {
	peerster.AreaChangeSession.Position = peerster.PathCar[1]
	peerster.AreaChangeSession.Active = true
	for {
		select {
		case <-peerster.AreaChangeSession.Channel:
			peerster.AreaChangeSession.Active = false
			break
		case <-time.After(6 * time.Second):
			peerster.AreaChangeSession.Channel <- true
		}
	}
}

func (peerster *Peerster) changeOfArea() bool {
	// If the area we are into is different to the one we are going, changing area
	if utils.AreaPositioner(peerster.PathCar[0]) != utils.AreaPositioner(peerster.PathCar[1]) {
		return true
	}
	return false
}
func (peerster *Peerster) positionAdvancer() {
	if peerster.collisionChecker() == false {
		peerster.PathCar = peerster.PathCar[1:]

		// There is a colision, do something
	} else {
		//If there has been more than 2 colision, negotiate
		if peerster.ColisionInfo.NumberColisions >= 2 {
			peerster.negotationOfColision()
			// If not just wait without moving to see if something changes
		} else {
			return
		}

	}
}
func (peerster *Peerster) collisionChecker() bool {
	peerster.PosCarsInArea.Mutex.Lock()
	defer peerster.PosCarsInArea.Mutex.Unlock()
	for _, carInfo := range peerster.PosCarsInArea.Slice {
		//If a car is in the position we want to move to, there is a collision
		if peerster.PathCar[1] == carInfo.Position {
			peerster.ColisionInfo.NumberColisions = peerster.ColisionInfo.NumberColisions + 1
			peerster.ColisionInfo.IPCar = carInfo.IPCar
			return true
		}
	}
	peerster.ColisionInfo.NumberColisions = 0
	peerster.ColisionInfo.IPCar = ""
	peerster.ColisionInfo.CoinFlip = 0
	return false
}

func negotiationCoinflip() int {

	min := 1
	max := 7000
	return rand.Intn(max-min+1) + min
}
func (peerster *Peerster) negotationOfColision() {
	// You flip a coin and send the information to the other guy
	//TODO: We have to add that if you are trying to change area,
	// and another guy from your current area wants to negotiate with you, you always win and stay still
	coinFlip := negotiationCoinflip()

	peerster.ColisionInfo.CoinFlip = coinFlip
	peerster.SendNegotiationMessage()

}
