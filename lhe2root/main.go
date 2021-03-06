// Convert LHE file into a ROOT TTree
package main

import (
	"flag"
	"fmt"
	"log"
	"io"
	"os"
	"strings"
	
	"go-hep.org/x/hep/lhef"	
	"go-hep.org/x/hep/groot"
	"go-hep.org/x/hep/groot/rtree"

	"github.com/rmadar/go-lorentz-vector/lv"
)

// Event stucture for partonic ttbar->dilepton event
type Event struct {	
	t    Particle
	tbar Particle
	b    Particle
	bbar Particle
	W    Particle
	Wbar Particle
	l    Particle
	lbar Particle
	v    Particle
	vbar Particle
}

type Particle struct {
	pt  float32
	eta float32
	phi float32
	m   float32
	pid int32
}

func main() {

	// Input arguments
	ifname  := flag.String("f", "ttbar_0j_parton.lhe", "Path to the input LHE file")
	tname   := flag.String("t", "truth", "Name of the created TTree")
	verbose := flag.Bool("v", false, "Enable verbose mode")
	flag.Parse()

	// Prepare the outfile and tree
	ofname := strings.ReplaceAll(*ifname, ".lhe", ".root")
	fout, err := groot.Create(ofname)
	if err != nil {
		log.Fatalf("could not create ROOT file %q: %w", ofname, err)
	}
	defer fout.Close()
	var e Event
	wvars := setBranches(&e)
	tw, err := rtree.NewWriter(fout, *tname, wvars)
	if err != nil {
		log.Fatalf("could not create scanner: %+v", err)
	}
	defer tw.Close()
	
	// Load LHE file
	f, err := os.Open(*ifname)
	if err != nil {
		panic(err)
	}

	// Get LHE decoder
	lhedec, err := lhef.NewDecoder(f)
	if err != nil {
		panic(err)
	}

	// Loop over events
	iEvt := 0
	for i := 0; ; i++ {

		// Decode this event, stop if the end of file is reached
		lhe_evt, err := lhedec.Decode()
		if err == io.EOF {
			break
		}

		// Print the event in verbose mode
		if *verbose {
			fmt.Println()
			fmt.Println(*lhe_evt)
		}
		
		// Converting the information from LHE event to TTree event
		var (
			pids     = lhe_evt.IDUP
			PxPyPzEM = lhe_evt.PUP
			get4Vec  = func(x [5]float64) (lv.FourVec) {
				return lv.NewFourVecPxPyPzM(x[0], x[1], x[2], x[4])
			}
			setPart  = func(part *Particle, P lv.FourVec, pid int64) {
				part.pt  = float32(P.Pt())
				part.eta = float32(P.Eta())
				part.phi = float32(P.Phi())
				part.m   = float32(P.M())
				part.pid = int32(pid)
			}
		)

		// Loop over particles
		for i, pid := range pids {
			if pid == 6 { // top-quark
				setPart(&e.t, get4Vec(PxPyPzEM[i]), pid)
			}
			if pid == -6 { // anti top-quark
				setPart(&e.tbar, get4Vec(PxPyPzEM[i]), pid)
			}
			if pid == 5 { // b-quark
				setPart(&e.b, get4Vec(PxPyPzEM[i]), pid)
			}
			if pid == -5 { // anti b-quark
				setPart(&e.bbar, get4Vec(PxPyPzEM[i]), pid)
			}
			if pid == 24 { // W+ boson
				setPart(&e.W, get4Vec(PxPyPzEM[i]), pid)
			}
			if pid == -24 { // W- boson
				setPart(&e.Wbar, get4Vec(PxPyPzEM[i]), pid)
			}
			if (pid == -11 || pid == -13 || pid == -15) { // Charged leptons
				setPart(&e.l, get4Vec(PxPyPzEM[i]), pid)
			}
			if (pid == 11 || pid == 13 || pid == 15) { // Charged anti-leptons
				setPart(&e.lbar, get4Vec(PxPyPzEM[i]), pid)
			}
			if (pid == 10 || pid == 12 || pid == 14) { // Neutrinos
				setPart(&e.v, get4Vec(PxPyPzEM[i]), pid)
			}
			if (pid == -10 || pid == -12 || pid == -14) { // Anti-neutrinos
				setPart(&e.vbar, get4Vec(PxPyPzEM[i]), pid)
			}			
		}
		
		// Write the TTree
		tw.Write()
		iEvt++
	}

	err = tw.Close()
	if err != nil {
		log.Fatalf("could not close tree-writer: %+v", err)
	}
	
	fmt.Println(" --> Event loop is done:", iEvt, "events processed and stored in", ofname)
}


func setBranches (e *Event) []rtree.WriteVar {
	return []rtree.WriteVar{

		// Top 
		{Name: "t_pt", Value: &e.t.pt},
		{Name: "t_eta", Value: &e.t.eta},
		{Name: "t_phi", Value: &e.t.phi},
		{Name: "t_pid", Value: &e.t.pid},
		{Name: "t_m", Value: &e.t.m},
		{Name: "tbar_pt", Value: &e.tbar.pt},
		{Name: "tbar_eta", Value: &e.tbar.eta},
		{Name: "tbar_phi", Value: &e.tbar.phi},
		{Name: "tbar_pid", Value: &e.tbar.pid},
		{Name: "tbar_m", Value: &e.tbar.m},

		// b-quarks
		{Name: "b_pt", Value: &e.b.pt},
		{Name: "b_eta", Value: &e.b.eta},
		{Name: "b_phi", Value: &e.b.phi},
		{Name: "b_pid", Value: &e.b.pid},
		{Name: "b_m", Value: &e.b.m},
		{Name: "bbar_pt", Value: &e.bbar.pt},
		{Name: "bbar_eta", Value: &e.bbar.eta},
		{Name: "bbar_phi", Value: &e.bbar.phi},
		{Name: "bbar_pid", Value: &e.bbar.pid},
		{Name: "bbar_m", Value: &e.bbar.m},

		// W-boson
		{Name: "W_pt", Value: &e.W.pt},
		{Name: "W_eta", Value: &e.W.eta},
		{Name: "W_phi", Value: &e.W.phi},
		{Name: "W_pid", Value: &e.W.pid},
		{Name: "W_m", Value: &e.W.m},
		{Name: "Wbar_pt", Value: &e.Wbar.pt},
		{Name: "Wbar_eta", Value: &e.Wbar.eta},
		{Name: "Wbar_phi", Value: &e.Wbar.phi},
		{Name: "Wbar_pid", Value: &e.Wbar.pid},
		{Name: "Wbar_m", Value: &e.Wbar.m},

		// Charged leptons
		{Name: "l_pt", Value: &e.l.pt},
		{Name: "l_eta", Value: &e.l.eta},
		{Name: "l_phi", Value: &e.l.phi},
		{Name: "l_pid", Value: &e.l.pid},
		{Name: "l_m", Value: &e.l.m},
		{Name: "lbar_pt", Value: &e.lbar.pt},
		{Name: "lbar_eta", Value: &e.lbar.eta},
		{Name: "lbar_phi", Value: &e.lbar.phi},
		{Name: "lbar_pid", Value: &e.lbar.pid},
		{Name: "lbar_m", Value: &e.lbar.m},

		// Neutrinos
		{Name: "v_pt", Value: &e.v.pt},
		{Name: "v_eta", Value: &e.v.eta},
		{Name: "v_phi", Value: &e.v.phi},
		{Name: "v_pid", Value: &e.v.pid},
		{Name: "v_m", Value: &e.v.m},
		{Name: "vbar_pt", Value: &e.vbar.pt},
		{Name: "vbar_eta", Value: &e.vbar.eta},
		{Name: "vbar_phi", Value: &e.vbar.phi},
		{Name: "vbar_pid", Value: &e.vbar.pid},
		{Name: "vbar_m", Value: &e.vbar.m},
	}
	
}
