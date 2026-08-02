package matchengine

import (
	"context"
	"reflect"
	"testing"
)

func TestCalibrationShortSampleIsDeterministicAndBounded(t *testing.T) {
	input := makeTestRoundInput(17001)
	left, err := CalibrateRounds(context.Background(), input, 48)
	if err != nil {
		t.Fatal(err)
	}
	right, err := CalibrateRounds(context.Background(), input, 48)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(left, right) {
		t.Fatalf("same calibration seed range differs: left=%+v right=%+v", left, right)
	}
	if left.Samples != 48 || left.AverageKills < 0 || left.AverageKills > 10 || left.AverageRoundDurationSeconds <= 0 {
		t.Fatalf("invalid calibration aggregate: %+v", left)
	}
	for name, rate := range map[string]float64{
		"t_win": left.TWinRate, "plant": left.PlantRate, "defuse": left.DefuseRate,
		"explosion": left.ExplosionRate, "first_kill": left.FirstKillWinnerRate,
		"five_v_three": left.FiveVThreeConversionRate, "three_v_five": left.ThreeVFiveComebackRate,
		"strong_team": left.StrongTeamWinRate,
	} {
		if rate < 0 || rate > 1 {
			t.Fatalf("%s rate is outside [0,1]: %f", name, rate)
		}
	}
}
