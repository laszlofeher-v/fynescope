#!/bin/bash
set -e
cd /home/lacidev/ai/fynescope

echo "=== Step 1: Rename files inside sim/ ==="
mv sim/sim.go sim/demo.go
mv sim/sim_test.go sim/demo_test.go
echo "Renamed sim/sim.go -> sim/demo.go"
echo "Renamed sim/sim_test.go -> sim/demo_test.go"

echo "=== Step 2: Rename gui/sim_gen.go ==="
mv gui/sim_gen.go gui/demo_gen.go
echo "Renamed gui/sim_gen.go -> gui/demo_gen.go"

echo "=== Step 3: Substitute in all .go files ==="
GO_FILES=$(find . -name "*.go" ! -path "./.git/*" ! -path "./vendor/*")

for f in $GO_FILES; do
    sed -i \
        -e 's|fynescope/sim"|fynescope/demo"|g' \
        -e 's|_ "fynescope/demo"|_ "fynescope/demo"|g' \
        -e 's|"sim/sim\.go"|"demo/demo.go"|g' \
        -e 's/^package sim$/package demo/' \
        -e 's/\bsim\.\(Default\|Min\|Max\|Set\|Get\|Interpolated\|FineGrained\)/demo.\1/g' \
        -e 's/sim\.DefaultChannels/demo.DefaultChannels/g' \
        -e 's/sim\.MinChannels/demo.MinChannels/g' \
        -e 's/sim\.MaxChannels/demo.MaxChannels/g' \
        -e 's/sim\.SetChannelCount/demo.SetChannelCount/g' \
        -e 's/sim\.SetNoiseAmplitude/demo.SetNoiseAmplitude/g' \
        -e 's/sim\.SetPhaseNoiseDegree/demo.SetPhaseNoiseDegree/g' \
        -e 's/sim\.SetRaiseFallTimePercent/demo.SetRaiseFallTimePercent/g' \
        -e 's/sim\.SetTriggerCalculationMode/demo.SetTriggerCalculationMode/g' \
        -e 's/sim\.InterpolatedTrigger/demo.InterpolatedTrigger/g' \
        -e 's/sim\.FineGrainedTrigger/demo.FineGrainedTrigger/g' \
        -e 's/\bOpenSimulator\b/OpenDemo/g' \
        -e 's/\bopenSimulator\b/openDemo/g' \
        -e 's/\bIsSimulator\b/IsDemo/g' \
        -e 's/\bsimulatorOnly\b/demoOnly/g' \
        -e 's/\bTestOpenSimulator_Success\b/TestOpenDemo_Success/g' \
        -e 's/\bTestOpenSimulator_SimulatorNotFound\b/TestOpenDemo_DemoNotFound/g' \
        -e 's/Simulator not found/Demo not found/g' \
        -e 's/\bSetSimGenMsg\b/SetDemoGenMsg/g' \
        -e 's/\bSetSimGenRsp\b/SetDemoGenRsp/g' \
        -e 's/\bSetSimRlcFilterMsg\b/SetDemoRlcFilterMsg/g' \
        -e 's/\bSetSimRlcFilterRsp\b/SetDemoRlcFilterRsp/g' \
        -e 's/\bSetSimGen\b/SetDemoGen/g' \
        -e 's/\bSetSimRlcFilter\b/SetDemoRlcFilter/g' \
        -e 's/\bSetSimGenCh\b/SetDemoGenCh/g' \
        -e 's/\bSimGenActiveTab\b/DemoGenActiveTab/g' \
        -e 's/\bSimGenPanel\b/DemoGenPanel/g' \
        -e 's/\bminSimGenElements\b/minDemoGenElements/g' \
        -e 's/simgenactivetab/demogenactivetab/g' \
        -e 's/simgenpanel/demogenpanel/g' \
        -e 's/\bSimId\b/DemoId/g' \
        -e 's/\bnewFfSimGenPanel\b/newFfDemoGenPanel/g' \
        -e 's/\bapplyFfSimGenSettings\b/applyFfDemoGenSettings/g' \
        -e 's/\bffSimGenPanel\b/ffDemoGenPanel/g' \
        -e 's/\bnewSimGenPanel\b/newDemoGenPanel/g' \
        -e 's/\bapplySimGenSettings\b/applyDemoGenSettings/g' \
        -e 's/"Simulator"/"Demo"/g' \
        -e 's/\bSimulator\b/Demo/g' \
        -e 's/\bsimulator\b/demo/g' \
        "$f"
done

echo "=== Step 4: Rename sim/ -> demo/ ==="
mv sim/ demo/
echo "Renamed sim/ -> demo/"

echo "=== Done! ==="
