"""A trivial tool to turn a string into a C++ constant.

This is not meant to be useful. It is only to provide an example of a tool that
generates code.
"""

import sys


def main(argv):
    if len(argv) != 4:
        raise Exception('usage: constant_generator out_file var_name text')
    with open(argv[1], 'w') as out:
        out.write('const char* %s = "%s";\n' % (argv[2], argv[3]))


if __name__ == '__main__':
    main(sys.argv)
