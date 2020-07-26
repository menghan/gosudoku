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
        for i, il in enumerate(self.array):
            for j, jv in enumerate(il):
                if jv == 0:
                    self.left += 1
                else:
                    self.avail[i][j] = set()
        self.recalc()

    def recalc(self):
        for i, il in enumerate(self.array):
            for j, jv in enumerate(il):
                self.recalc_pos(i, j)

    def recalc_pos(self, i, j):
        v = self.array[i][j]
        if v == 0:
            return
        for ii in xrange(9):
            self.avail[ii][j].discard(v)
        for jj in xrange(9):
            self.avail[i][jj].discard(v)
        sqi, sqj = i / 3 * 3, j / 3 * 3
        for ii in xrange(sqi, sqi + 3):
            for jj in xrange(sqj, sqj + 3):
                self.avail[ii][jj].discard(v)

    def only_possible_solution(self):
        for i, ia in enumerate(self.avail):
            for j, ja in enumerate(ia):
                if self.array[i][j] == 0 and len(ja) == 1:
                    return i, j, list(ja)[0]

    def pair_solution(self):

        def search_for_duplicated(avs):
            if len(set(avs)) == len(avs):
                return
            avs = list(avs)
            avs.sort()
            for i, av in enumerate(avs):
                if i > 0 and avs[i-1] == avs[i]:
                    if len(av) == 3:
                        try:
                            if avs[i-2] == avs[i]:
                                return av
                            if avs[i+1] == avs[i]:
                                return av
                        except IndexError:
                            pass
                    if len(av) == 2:
                        return av

        # line
        for i, ia in enumerate(self.avail):
            avs = []
            for j, ja in enumerate(ia):
                avs.append(tuple(sorted(ja)))

            v = search_for_duplicated(avs)
            if v is None:
                continue
            for j, ja in enumerate(ia):
                if tuple(sorted(ja)) == v or self.array[i][j] != 0:
                    continue
                for x in v:
                    ja.discard(x)
                if len(ja) == 1:
                    return i, j, list(ja)[0], [(i, jj) for jj in xrange(9) if tuple(sorted(self.avail[i][jj])) == v], list(v)

        # column
        for j, ja in enumerate(self.avail):
            avs = []
            for i in xrange(9):
                ia = self.avail[i][j]
                avs.append(tuple(sorted(ia)))

            v = search_for_duplicated(avs)
            if v is None:
                continue
            for i in xrange(9):
                ia = self.avail[i][j]
                if tuple(sorted(ia)) == v or self.array[i][j] != 0:
                    continue
                for x in v:
                    ia.discard(x)
                if len(ia) == 1:
                    return i, j, list(ia)[0], [(ii, j) for ii in xrange(9) if tuple(sorted(self.avail[ii][j])) == v], list(v)

    def need_to_be_solution(self):
        # line
        for i, ia in enumerate(self.array):
            s = set(range(1, 10))
            for j, v in enumerate(ia):
                if v != 0:
                    s.discard(v)
            for needto in s:
                can_fill_needto = [j for j, av in enumerate(self.avail[i]) if self.array[i][j] == 0 and needto in av]
                if len(can_fill_needto) == 1:
                    return i, can_fill_needto[0], needto, '行'

        # column
        for j, ja in enumerate(self.array):
            s = set(range(1, 10))
            for i, v in enumerate(ja):
                if v != 0:
                    s.discard(v)
            for needto in s:
                can_fill_needto = [i for i in xrange(9) if self.array[i][j] == 0 and needto in self.avail[i][j]]
                if len(can_fill_needto) == 1:
                    return can_fill_needto[0], j, needto, '列'

        # square
        for sqi in [0, 3, 6]:
            for sqj in [0, 3, 6]:
                s = set(range(1, 10))
                for ii in xrange(sqi, sqi + 3):
                    for jj in xrange(sqj, sqj + 3):
                        s.discard(self.array[ii][jj])
                for needto in s:
                    can_fill_needto = []
                    for i in xrange(sqi, sqi + 3):
                        for j in xrange(sqj, sqj + 3):
                            if self.array[i][j] == 0 and needto in self.avail[i][j]:
                                can_fill_needto.append((i, j))
                    if len(can_fill_needto) == 1:
                        return can_fill_needto[0][0], can_fill_needto[0][1], needto, '框'

    def fill(self, x, y, v):
        self.array[x][y] = v
        self.avail[x][y] = set()
        self.recalc_pos(x, y)
        self.left -= 1

    def printto(self, out):
        print >> out, '   ',
        for i in xrange(9):
            print >> out, '%d ' % (i + 1),
        print >> out
        print >> out, '    -------------------------'
        for i, il in enumerate(self.array):
            print >> out, '%d: ' % (i + 1),
            for j, jv in enumerate(il):
                if jv == 0:
                    print >> out, '_ ',
                else:
                    print >> out, '%d ' % jv,
            print >> out


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
    s = read_sudoku(open(sys.argv[1]).read())
    s.printto(sys.stdout)
    print '共 %d 空' % s.left
    sys.stdout.flush()
    raw_input('Press Enter to Continue...\n')
    while s.left:
        pos = s.need_to_be_solution()
        if pos is not None:
            x, y, v, hint = pos
            print '第 %d 行 第 %d 列经 %s必须法 可以填 %d' % (x + 1, y + 1, hint, v)
            sys.stdout.flush()
            s.fill(x, y, v)
            s.printto(sys.stdout)
            sys.stdout.flush()
            raw_input('Press Enter to Continue...\n')
            continue

        pos = s.only_possible_solution()
        if pos is not None:
            x, y, v = pos
            print '第 %d 行 第 %d 列经 排除法 可以填 %d' % (x + 1, y + 1, v)
            sys.stdout.flush()
            s.fill(x, y, v)
            s.printto(sys.stdout)
            sys.stdout.flush()
            raw_input('Press Enter to Continue...\n')
            continue

        pos = s.pair_solution()
        if pos is not None:
            x, y, v, pairs, pairs_value = pos
            print '第 %d 行 第 %d 列经 成对成组法:[%s]填%s 可以填 %d' % (x + 1, y + 1, ','.join(['%s行%s列' % (i+1,j+1) for (i,j) in pairs]), ','.join([str(vv) for vv in pairs_value]), v)
            sys.stdout.flush()
            s.fill(x, y, v)
            s.printto(sys.stdout)
            sys.stdout.flush()
            raw_input('Press Enter to Continue...\n')
            continue

        print '没找到方法, 还剩余 %d 空' % s.left
        sys.stdout.flush()
        break

    if s.left == 0:
        print '完成'
        sys.stdout.flush()

if __name__ == '__main__':
    main()
