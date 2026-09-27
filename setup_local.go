//go:build !remote

package main

// In a local build the controller is created inside gui.ScpDesc.Menu() by calling
// control.NewControl(con).  No extra setup is needed.
