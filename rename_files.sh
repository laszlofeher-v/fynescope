#!/bin/bash

# This script renames files across the project to use snake_case
# for better readability, addressing ambiguous concatenated names like "channellabel".

echo "Renaming trigger components..."
mv gui/advtriggerpoint.go gui/adv_trigger_point.go
mv gui/complextrigger.go gui/complex_trigger.go
mv gui/triggerpoint.go gui/trigger_point.go

echo "Renaming channel labels..."
mv gui/ftchannellabel.go gui/ft_channel_label.go
mv gui/dftchannellabel.go gui/dft_channel_label.go

echo "Renaming rasters..."
mv gui/dftraster.go gui/dft_raster.go
mv gui/ffraster.go gui/ff_raster.go
mv gui/ffraster_test.go gui/ff_raster_test.go
mv gui/ftraster.go gui/ft_raster.go
mv gui/fvraster.go gui/fv_raster.go

echo "Renaming generator controls..."
mv gui/extgen.go gui/ext_gen.go
mv gui/simgen.go gui/sim_gen.go

echo "Renaming other GUI components..."
mv gui/screendraw.go gui/screen_draw.go
mv gui/testproxy.go gui/test_proxy.go
mv gui/timediv.go gui/time_div.go

echo "Renaming control package files..."
mv control/blockmode.go control/block_mode.go
mv control/extgen.go control/ext_gen.go
mv control/screentime.go control/screen_time.go
mv control/screentime_test.go control/screen_time_test.go

echo "Renaming psc package files..."
mv psc/psconsts.go psc/ps_consts.go

echo "Renaming ps2000a package files..."
mv ps2000a/noscope.go ps2000a/no_scope.go

echo "Renaming custom widget packages..."
mv checkcolorpick/checkcolorpick.go checkcolorpick/check_color_pick.go
mv checkcolorpick/checkcolorpick_test.go checkcolorpick/check_color_pick_test.go
mv sliderscroll/sliderscroll.go sliderscroll/slider_scroll.go
mv sliderscroll/sliderscroll_test.go sliderscroll/slider_scroll_test.go
mv tastybutton/tastybutton.go tastybutton/tasty_button.go
mv tastybutton/tastybutton_test.go tastybutton/tasty_button_test.go

echo "Done! Remember to run 'go test ./...' to ensure everything still builds."
