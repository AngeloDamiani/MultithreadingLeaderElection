package LeaderElection

import "fmt"

type Identifier uint16
const StarterID = 42

type LStatus int
const (
	LEADER   LStatus = 0
	FOLLOWER LStatus = 1
	CANDIDATE LStatus = 2
)

type Message struct {
	Sender Identifier
	Leader Identifier
}

type NetworkNode struct {
	ID        Identifier
	Neighbors map[Identifier]*chan Message
	MailBox   *chan Message
	Status    LStatus
	Leader    Identifier

	MasterMailbox *chan Message
	printMutex *chan int
}

func (v* NetworkNode) Init(id Identifier, mailbox *chan Message, mutex *chan int, master *chan Message){
	v.ID = id
	v.Leader = id
	v.MailBox = mailbox
	v.Status = CANDIDATE
	v.Neighbors = make(map[Identifier]*chan Message)

	v.MasterMailbox = master
	v.printMutex = mutex
}

func (v* NetworkNode) sendKnowledge(neighbor Identifier){
	select {
		case *v.Neighbors[neighbor] <- Message{v.ID,v.Leader}:
		default:
	}
}

func (v* NetworkNode) notifyMaster(){
	*v.MasterMailbox <- Message{v.ID,v.Leader}
}

func (v* NetworkNode) rcvKnowledge()(msg Message){
	return <- *v.MailBox
}

func (v* NetworkNode) addNeighbor(nodeID Identifier, nodeMailbox *chan Message){
	v.Neighbors[nodeID] = nodeMailbox
}

func Connect(nodeA *NetworkNode, nodeB *NetworkNode){
	nodeA.addNeighbor(nodeB.ID, nodeB.MailBox)
	nodeB.addNeighbor(nodeA.ID, nodeA.MailBox)
}

func (v *NetworkNode) floodKnowledge(){
	for id, _ := range v.Neighbors{
		v.sendKnowledge(id)
	}

	v.notifyMaster()
}

func (v *NetworkNode) propagateKnowledge(msg Message){
	for id, _ := range v.Neighbors{
		if  id!=msg.Sender{
			v.sendKnowledge(id)
		}
	}

	v.notifyMaster()
}

func (v *NetworkNode) imLeader(){
	v.Leader = v.ID
	v.Status = LEADER
}

func (v *NetworkNode) printKnowledge(){
	status := ""
	switch v.Status {
	case LEADER:
		status = "Leader"
	case FOLLOWER:
		status = "Follower"
	case CANDIDATE:
		status = "Candidate"
	}

	<- *v.printMutex
	fmt.Print("[" )
	fmt.Print(v.ID)
	fmt.Print("][" + status +"]: My Leader is: ")
	fmt.Println(v.Leader)
	*v.printMutex <- 1
}

func RunNode(node *NetworkNode){
	if node.ID == StarterID{
		node.imLeader()
		node.printKnowledge()
		node.floodKnowledge()
	}
	for{
		msg := node.rcvKnowledge()

		// Exit condition, convergence happened
		if msg.Sender == 0{
			break
		}

		if msg.Leader>node.Leader{
			// Case node's knowledge is better
			if node.Status == CANDIDATE {
				node.imLeader()
				node.printKnowledge()
				node.floodKnowledge()

			} else {
				node.sendKnowledge(msg.Sender)
			}

		} else {

			if msg.Leader != node.Leader {
				// Case node's knowledge must be updated
				node.Status = FOLLOWER
				node.Leader = msg.Leader
				node.printKnowledge()
				node.propagateKnowledge(msg)
			}
		}
	}
}
