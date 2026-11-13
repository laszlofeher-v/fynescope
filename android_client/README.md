# Fynescope Android Client (`android_client`)

An Android client for the headless `fynescope-server` (`cmd/fynescope-server`), adapted for small landscape screens with vertically scrollable control tabs.

## UI Adaptations for Small Screens
- **Landscape Orientation**: Locked to landscape mode in `AndroidManifest.xml` (`android:screenOrientation="landscape"`).
- **Vertically Scrollable Control Tabs**: All oscilloscope control tabs (`f(t)`, `f(v)`, `FFT`, `f(f)`, `RLC`, `digital`, `gen`, `extgen`, `digGen`, `vch`, `decode`, `filter`, `corr`, `multiscope`) are wrapped in vertical scroll containers so all channel, timebase, trigger, generator, and analysis controls are accessible on compact screens.
- **Compact Mobile Toolbar & Server Reconnection**:
  - Top toolbar is horizontally scrollable (`container.NewHScroll`) and hides desktop-only window buttons (Fullscreen/Restore/Undock).
  - Includes a **Server** button (`HomeIcon`) to disconnect cleanly and return to the Server Connection screen at any time.
- **Touch Support**:
  - Touch-dragging waveforms, channel offset labels, and trigger markers works directly on the oscilloscope display.
  - Tapping the upper or lower half of a digit on 7-segment displays (`disp7`) increments or decrements that digit on mobile devices.

## Building an Android APK

From `android_client/`:

```bash
cd android_client
ANDROID_HOME=$HOME/Android/Sdk ANDROID_NDK_HOME=$HOME/android-ndk-r25c \
  fyne package -os android -app-id com.github.laszlofeher.fynescope.android -icon ../Icon.png
```

## Running Locally (Desktop Preview of Mobile UI)

1. Start the headless server in demo mode:
   ```bash
   go run ./cmd/fynescope-server -demo -listen :9876
   ```
2. Run the Android client UI locally:
   ```bash
   go run ./android_client -addr 127.0.0.1:9876 -connect
   ```

