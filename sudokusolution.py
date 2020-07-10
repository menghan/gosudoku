#!/usr/bin/env python
# coding=utf-8

import sys


class Sudoku(object):

    def __init__(self, array):
        self.array = array
        self.avail = [
            [set(range(1, 10)) for i in xrange(9)]
            for j in xrange(9)
        ]
        self.left = 0
        self.recalc()

    def recalc(self):
        for i, il in enumerate(self.array):
            for j, jv in enumerate(il):
                self.recalc_pos(i, j)

    def recalc_pos(self, i, j):
        v = self.array[i][j]
        if v == 0:
            self.left += 1
            return
        for ii in xrange(9):
            self.avail[ii][j].discard(v)
        for jj in xrange(9):
            self.avail[i][jj].discard(v)
        sqi, sqj = i / 3 * 3, j / 3 * 3
        for ii in xrange(sqi, sqi + 3):
            for jj in xrange(sqj, sqj + 3):
                self.avail[ii][jj].discard(v)

    def easy_position(self):
        for i, ia in enumerate(self.avail):
            for j, ja in enumerate(ia):
                if len(ja) == 1:
                    assert self.array[i][j] == 0  # FIXME
                    return i, j, list(ja)[0]

    def fill(self, x, y, v):
        self.array[x][y] = v
        self.recalc_pos(x, y)
        self.left -= 1


def read_sudoku(body):
    array = [[0 for i in xrange(9)] for j in xrange(9)]
    for i, line in enumerate(body.splitlines()[:9]):
        for j, c in enumerate(line[:9]):
            if c in (' ', '_'):
                array[i][j] = 0
            else:
                v = int(c)
                assert 1 <= v <= 9
                array[i][j] = v
    return Sudoku(array)


def main():
    s = read_sudoku(sys.stdin.read())
    print s.array
    while s.left:
        pos = s.easy_position()
        if pos is not None:
            x, y, v = pos
            print '第 %d 行 第 %d 列经排除可以填 %d' % (x + 1, y + 1, v)
            sys.stdout.flush()
            s.fill(x, y, v)
            continue

        assert False, '没找到简单排除, 还剩余 %d 空' % s.left

    print '完成'
    sys.stdout.flush()

if __name__ == '__main__':
    main()
