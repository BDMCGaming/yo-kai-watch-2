package matchmake_extension

import (
	//"fmt"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	commonglobals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"
	"github.com/PretendoNetwork/nex-protocols-common-go/v2/matchmake-extension/database"
	matchmakeextension "github.com/PretendoNetwork/nex-protocols-go/v2/matchmake-extension"
	"github.com/PretendoNetwork/yo-kai-watch-2/globals"
)

// Borrowed code from PUYOPUYOTETRIS. Yo-Kai Watch 2 has the same quirk of having everyone send CloseParticipation
// This does seem to cause issues on the player side however, as one player will freeze on the results screen due to an unhandled error

func CloseParticipation(err error, packet nex.PacketInterface, callID uint32, gid types.UInt32) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		commonglobals.Logger.Error(err.Error())
		return nil, nex.NewError(nex.ResultCodes.Core.InvalidArgument, "change_error")
	}

	connection := packet.Sender().(*nex.PRUDPConnection)
	manager := globals.MatchmakingManager
	endpoint := connection.Endpoint().(*nex.PRUDPEndPoint)

	manager.Mutex.Lock()

	session, _, nexError := database.GetMatchmakeSessionByID(manager, endpoint, uint32(gid))
	if nexError != nil {
		manager.Mutex.Unlock()
		return nil, nexError
	}

	// * Yo-Kai Watch 2 has both players in the match send CloseParticipation
	// * So, if a non-owner asks, just lie and claim success without actually changing anything.
	if session.Gathering.OwnerPID.Equals(connection.PID()) {
		nexError = database.UpdateParticipation(manager, uint32(gid), false)
		if nexError != nil {
			manager.Mutex.Unlock()
			return nil, nexError
		}
	}

	manager.Mutex.Unlock()

	rmcResponse := nex.NewRMCSuccess(endpoint, nil)
	rmcResponse.ProtocolID = matchmakeextension.ProtocolID
	rmcResponse.MethodID = matchmakeextension.MethodCloseParticipation
	rmcResponse.CallID = callID

	return rmcResponse, nil
}
