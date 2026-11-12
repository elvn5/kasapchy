#!/usr/bin/env python3
"""Validate PT_LOAD alignment of every native ELF library in an Android bundle."""
import argparse
import struct
import zipfile

parser = argparse.ArgumentParser()
parser.add_argument('bundle')
args = parser.parse_args()
count = 0
with zipfile.ZipFile(args.bundle) as bundle:
    for name in bundle.namelist():
        if not name.endswith('.so'):
            continue
        data = bundle.read(name)
        if data[:4] != b'\x7fELF':
            raise SystemExit(f'Not ELF: {name}')
        endian = '<' if data[5] == 1 else '>'
        is64 = data[4] == 2
        phoff = struct.unpack_from(endian + ('Q' if is64 else 'I'), data, 32 if is64 else 28)[0]
        entsize, num = struct.unpack_from(endian + 'HH', data, 54 if is64 else 42)
        aligns = []
        for index in range(num):
            offset = phoff + index * entsize
            if struct.unpack_from(endian + 'I', data, offset)[0] == 1:
                align = struct.unpack_from(endian + ('Q' if is64 else 'I'), data, offset + (48 if is64 else 28))[0]
                aligns.append(align)
                if align < 16384:
                    raise SystemExit(f'FAIL {name}: PT_LOAD alignment {align}')
        if not aligns:
            raise SystemExit(f'No PT_LOAD segments: {name}')
        count += 1
        print(f'OK {name}: {aligns}')
print(f'Checked {count} ELF libraries. Also validate BundleConfig and generated APK zip alignment.')
