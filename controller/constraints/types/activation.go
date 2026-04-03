package types

type Activation int

const Yes Activation = 1
const No Activation = -1
const NoOpinion Activation = 0

func (a Activation) String() string {
	switch a {
	case No:
		return "No"
	case Yes:
		return "Yes"
	case NoOpinion:
		return "NoOpinion"
	}
	return "Unknown"
}

var orMatrix = [3][3]Activation{
	{No, No, Yes},
	{No, NoOpinion, Yes},
	{Yes, Yes, Yes},
}

func (a Activation) Or(b Activation) Activation {
	return orMatrix[a+1][b+1]
}

var andMatrix = [3][3]Activation{
	{No, No, No},
	{No, NoOpinion, Yes},
	{No, Yes, Yes},
}

func (a Activation) And(b Activation) Activation {
	return andMatrix[a+1][b+1]
}

func (a Activation) Not() Activation {
	return -a
}
