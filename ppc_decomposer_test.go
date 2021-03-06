package gapstone

import "testing"
import "bytes"
import "fmt"
import "io/ioutil"

func ppcInsnDetail(insn Instruction, engine *Engine, buf *bytes.Buffer) {

	if len(insn.PPC.Operands) > 0 {
		fmt.Fprintf(buf, "\top_count: %v\n", len(insn.PPC.Operands))
	}

	for i, op := range insn.PPC.Operands {
		switch op.Type {
		case PPC_OP_REG:
			fmt.Fprintf(buf, "\t\toperands[%v].type: REG = %v\n", i, engine.RegName(op.Reg))
		case PPC_OP_IMM:
			fmt.Fprintf(buf, "\t\toperands[%v].type: IMM = 0x%x\n", i, (uint64(op.Imm)))
		case PPC_OP_MEM:
			fmt.Fprintf(buf, "\t\toperands[%v].type: MEM\n", i)
			if op.Mem.Base != PPC_REG_INVALID {
				fmt.Fprintf(buf, "\t\t\toperands[%v].mem.base: REG = %s\n",
					i, engine.RegName(op.Mem.Base))
			}
			if op.Mem.Disp != 0 {
				fmt.Fprintf(buf, "\t\t\toperands[%v].mem.disp: 0x%x\n", i, uint64(op.Mem.Disp))
			}
		}

	}

	if insn.PPC.BC != 0 {
		fmt.Fprintf(buf, "\tBranch code: %v\n", insn.PPC.BC)
	}

	if insn.PPC.BH != 0 {
		fmt.Fprintf(buf, "\tBranch hint: %v\n", insn.PPC.BH)
	}

	if insn.PPC.UpdateCR0 {
		fmt.Fprintf(buf, "\tUpdate-CR0: True\n")
	}

	fmt.Fprintf(buf, "\n")
}

func TestPPC(t *testing.T) {

	final := new(bytes.Buffer)
	spec_file := "ppc.SPEC"

	for i, platform := range ppc_tests {

		engine, err := New(platform.arch, platform.mode)
		if err != nil {
			t.Errorf("Failed to initialize engine %v", err)
			return
		}
		for _, opt := range platform.options {
			engine.SetOption(opt.ty, opt.value)
		}
		if i == 0 {
			maj, min := engine.Version()
			t.Logf("Arch: PPC. Capstone Version: %v.%v", maj, min)
		}
		defer engine.Close()

		insns, err := engine.Disasm([]byte(platform.code), address, 0)
		if err == nil {
			fmt.Fprintf(final, "****************\n")
			fmt.Fprintf(final, "Platform: %s\n", platform.comment)
			fmt.Fprintf(final, "Code:")
			dumpHex([]byte(platform.code), final)
			fmt.Fprintf(final, "Disasm:\n")
			for _, insn := range insns {
				fmt.Fprintf(final, "0x%x:\t%s\t%s\n", insn.Address, insn.Mnemonic, insn.OpStr)
				ppcInsnDetail(insn, &engine, final)
			}
			fmt.Fprintf(final, "0x%x:\n", insns[len(insns)-1].Address+insns[len(insns)-1].Size)
			fmt.Fprintf(final, "\n")
		} else {
			t.Errorf("Disassembly error: %v\n", err)
		}

	}

	spec, err := ioutil.ReadFile(spec_file)
	if err != nil {
		t.Errorf("Cannot read spec file %v: %v", spec_file, err)
	}
	if fs := final.String(); string(spec) != fs {
		fmt.Println(fs)
		t.Errorf("Output failed to match spec!")
	} else {
		t.Logf("Clean diff with %v.\n", spec_file)
	}

}
