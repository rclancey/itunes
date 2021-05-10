#!/usr/bin/env python

import sys
import os
import json
import re
import subprocess

cpat = re.compile('[^A-Za-z0-9_]')

def make_enum(name, values):
    out = 'type %s int\n\n' % name
    out += 'var %sNames = map[%s]string{\n' % (name, name)
    for k, v in sorted(values.items(), key=lambda (k, v): v):
        out += '\t%s(0x%X): "%s",\n' % (name, v, k)
    out += '}\n'
    out += 'var %sValues = map[string]%s{\n' % (name, name)
    for k, v in sorted(values.items(), key=lambda (k, v): v):
        out += '\t"%s": %s(0x%X),\n' % (k, name, v)
    out += '}\n'
    out += 'const (\n'
    for k, v in sorted(values.items(), key=lambda (k, v): v):
        ck = cpat.sub('', k.upper())
        out += '\t%s_%s = %s(0x%X)\n' % (name, ck, name, v)
    out += ')\n\n'
    out += 'func (e %s) String() string{\n' % name
    out += '\ts, ok := %sNames[e]\n' % name
    out += '\tif ok {\n'
    out += '\t\treturn s\n'
    out += '\t}\n'
    out += '\treturn fmt.Sprintf("%s_0x%%X", int(e))\n' % name
    out += '}\n\n'
    out += 'func (e %s) MarshalJSON() ([]byte, error) {\n' % name
    out += '\treturn json.Marshal(e.String())\n'
    out += '}\n\n'
    out += 'func (e *%s) UnmarshalJSON(data []byte) error {\n' % name
    out += '\tvar s string\n'
    out += '\terr := json.Unmarshal(data, &s)\n'
    out += '\tif err != nil {\n'
    out += '\t\treturn err\n'
    out += '\t}\n'
    out += '\tv, ok := %sValues[s]\n' % name
    out += '\tif !ok {\n'
    out += '\t\treturn fmt.Errorf(\"unknown %s %%s\", s)\n' % name
    out += '\t}\n'
    out += '\t*e = v\n'
    out += '\treturn nil\n'
    out += '}\n\n'
    return out

def main():
    infn = os.path.join(os.path.dirname(__file__), 'enums.json')
    f = open(infn, 'r')
    enums = json.load(f)
    f.close()
    outfn = os.path.join(os.path.dirname(__file__), 'enums.go')
    out = 'package itunes\n\n'
    out += 'import (\n'
    out += '\t"encoding/json"\n'
    out += '\t"fmt"\n'
    out += ')\n\n'
    for name, values in sorted(enums.items()):
        out += make_enum(name, values)
    f = open(outfn, 'w')
    p = subprocess.Popen(['gofmt'], stdin=subprocess.PIPE, stdout=f)
    p.stdin.write(out.encode('utf-8'))
    p.stdin.close()
    p.wait()
    f.close()

if '__main__' == __name__:
    main()
