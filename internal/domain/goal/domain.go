package goal

import "time"

// GoalIcon is a decorative icon for a financial goal.
type GoalIcon string

const (
	IconLove      GoalIcon = "love"
	IconEmergency GoalIcon = "emergency"
	IconVehicle   GoalIcon = "vehicle"
	IconHome      GoalIcon = "home"
	IconTravel    GoalIcon = "travel"
	IconOther     GoalIcon = "other"
)

// GoalStatus represents the lifecycle state of a goal.
type GoalStatus string

const (
	GoalStatusActive    GoalStatus = "active"
	GoalStatusCompleted GoalStatus = "completed"
	GoalStatusArchived  GoalStatus = "archived"
)

// Goal is the core financial goal entity.
type Goal struct {
	ID            string
	UserID        string
	Title         string
	Icon          *GoalIcon
	TargetAmount  float64
	CurrentAmount float64
	MonthlyTarget *float64
	Currency      string
	TargetDate    *time.Time
	Status        GoalStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ProgressPercent returns what percentage of the target has been saved.
// Returns 100.0 when TargetAmount is zero to avoid division by zero.
func (g *Goal) ProgressPercent() float64 {
	if g.TargetAmount <= 0 {
		return 100.0
	}
	pct := (g.CurrentAmount / g.TargetAmount) * 100
	if pct > 100 {
		return 100.0
	}
	return pct
}

// Remaining returns how much more needs to be saved to reach the target.
// Returns 0 when the goal is already met or exceeded.
func (g *Goal) Remaining() float64 {
	r := g.TargetAmount - g.CurrentAmount
	if r < 0 {
		return 0
	}
	return r
}
