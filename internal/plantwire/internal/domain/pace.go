package domain

const ClockModeWall = "wall"

type WallSchedule struct {
	WorldStartNS int64
	WallStartNS  int64
}

func WallScheduleFrom(wallNowNS int64) WallSchedule {
	start := wallNowNS - wallNowNS%1_000_000
	return WallSchedule{WorldStartNS: start, WallStartNS: start}
}

func (s WallSchedule) WallAt(worldNS int64) int64 {
	return s.WallStartNS + (worldNS - s.WorldStartNS)
}

func (s WallSchedule) WorldAt(wallNS int64) int64 {
	return s.WorldStartNS + (wallNS - s.WallStartNS)
}

func AttachedSlice(clockNS, sliceNS, worldNowNS int64) int64 {
	if sliceNS <= 0 || worldNowNS <= clockNS {
		return clockNS
	}
	return clockNS + (worldNowNS-clockNS)/sliceNS*sliceNS
}

func (s WallSchedule) WaitNS(worldNS, wallNowNS int64) int64 {
	return max(s.WallAt(worldNS)-wallNowNS, 0)
}
