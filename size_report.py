#!/usr/bin/env python3

# Do a profiling build:
# enable_nacl=false
# enable_profiling=true
# is_component_build=false
# is_debug=false
# symbol_level=2

# Then generate this file:
# nm -ClSp out/Release/chrome > binary.txt

import collections
import itertools
import math
import re

def escape(s):
  return s.replace('\\', '\\\\').replace('\'', '\\\'').replace('"', r'\\"')

def prefix(a, b):
  i = 0
  for (u, v) in zip(a, b):
    if u != v:
      break
    i += 1
  return a[0:i]


def aggregate(longest_prefix_len, a, b):
  aggregate_name = a.name[0:longest_prefix_len]
  a.name = a.name[longest_prefix_len:]
  b.name = b.name[longest_prefix_len:]
  return Aggregate(aggregate_name, [a, b])


SUFFIXES = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB']
def size_to_str(n):
  order = int(math.log2(n) / 10)
  return '%d %s' % (int(0.5 + n / math.pow(2, order * 10)), SUFFIXES[order])


class Node(object):
  def __init__(self):
    self.suffix = None

  # The ID of the node. Note that this is not stable if the name, suffix or
  # size changes.
  @property
  def node_id(self):
    s = '%s %s' % (size_to_str(self.size), '/'.join(self.name))
    if self.suffix:
      s += '#%d' % self.suffix
    return s


class Aggregate(Node):
  def __init__(self, name, parts):
    super(Aggregate, self).__init__()
    self.name = name
    self.parts = parts

  def insert(self, entry, longest_prefix_len=None):
    if longest_prefix_len is None:
      longest_prefix_len = len(prefix(self.name, entry.name))
    if longest_prefix_len < len(self.name):
      return aggregate(longest_prefix_len, self, entry)
    # Trim the common part of the name.
    entry.name = entry.name[longest_prefix_len:]
    longest_prefix_len = None
    # Does a binary search for where to insert |entry|
    def find_insertion_point(lo, hi):
      if lo == hi:
        return lo
      mid = lo + (hi - lo) // 2
      if entry.name < self.parts[mid].name:
        return find_insertion_point(lo, mid)
      else:
        return find_insertion_point(mid + 1, hi)
    n = find_insertion_point(0, len(self.parts))
    # Inspect the previous, next entry to see which has more in common
    if n == 0:
      longest_prefix_len = 0
      longest_prefix_entry = 0
    else:
      longest_prefix_len = len(prefix(self.parts[n-1].name, entry.name))
      longest_prefix_entry = n - 1
    if n < len(self.parts):
      prefix_len = len(prefix(self.parts[n].name, entry.name))
      if prefix_len > longest_prefix_len:
        longest_prefix_len = prefix_len
        longest_prefix_entry = n
    if longest_prefix_len == 0:
      # Nothing in common: New node.
      self.parts.insert(n, entry)
      return self
    else:
      # Merge the new entry with the child at that point
      self.parts[longest_prefix_entry] = (
          self.parts[longest_prefix_entry].insert(entry, longest_prefix_len))
      return self

  def dump(self, level=0):
    print(' ' * level, self.node_id)
    for part in self.parts:
      part.dump(level + 2)

  def preorder(self):
    yield self
    yield from itertools.chain.from_iterable(
        map(lambda part: part.preorder(), self.parts))

  def update_stats(self):
    self.size = sum(map(lambda part: part.update_stats(), self.parts))
    return self.size

  def write_report(self, f, parent):
    f.write('["%s", %s, %d, %d],\n' % (
        escape(self.node_id),
        parent is None and 'null' or '"%s"' % escape(parent.node_id),
        0,
        0))
    for part in self.parts:
      part.write_report(f, self)


class Entry(Node):
  def __init__(self, size, path, symbol):
    super(Entry, self).__init__()
    self.size = size
    self.path = path
    self.symbol = symbol
    self.name = path.split('/') + (symbol and [symbol] or [])

  def insert(self, other, longest_prefix_len=None):
    if other.path == self.path and other.symbol == self.symbol:
      self.size += other.size
      return self
    if longest_prefix_len is None:
      longest_prefix_len = len(prefix(self.name, other.name))
    return aggregate(longest_prefix_len, self, other)

  def dump(self, level=0):
    print(' ' * level, self.node_id)

  def preorder(self):
    yield self

  def update_stats(self):
    return self.size

  def write_report(self, f, parent):
    f.write('["%s", %s, %d, %d],\n' % (
        escape(self.node_id),
        parent is None and 'null' or '"%s"' % escape(parent.node_id),
        self.size,
        0))


# Walks a tree ensuring nodes have unique IDs.
def uniquify(tree):
  ids = collections.defaultdict(int)
  # It's a quirk of Google Charts TreeMap that it treats '' as null.
  ids[''] = 1
  for node in tree.preorder():
    ids[node.node_id] += 1
    n = ids[node.node_id]
    if n > 1:
      node.suffix = n

# TODO: A lot of apparently empty names, work out how the tree building is broken.

class Summarizer(object):
  def __init__(self):
    self._sized = re.compile('^[0-9a-f]{16} (?P<size>[0-9a-z]{16}) (?P<section>[a-zA-Z]) (?P<symbol>[^\t\n]+)(\t(?P<path>[^:]+):(?P<line>[0-9]+))?')
    self._matches = 0
    self._lines = 0
    self._entries = []
    self._tree = None

  def process_line(self, line):
    self._lines += 1
    m = self._sized.match(line)
    if not m:
      return
    self._matches += 1
    size = int(m.group('size'), 16)
    symbol = m.group('symbol')
    path = m.group('path')
    if not path:
      # This is a compromise between utility and output size. Chrome can't
      # render ginormous treemaps.
      path = symbol[0:10] + '...'
    # Pass symbol here to get symbol-level breakdowns.
    # This drastically increases the size of the output file.
    entry = Entry(size, path, None)
    self._entries.append(entry)

  def end(self):
    self._tree = self._entries[0]
    for entry in self._entries[1:]:
      self._tree = self._tree.insert(entry)
    self._tree.update_stats()
    uniquify(self._tree)

  def write_report(self, filename):
    marker = 'FIXME'
    i = TEMPLATE.find(marker)
    with open(filename, 'w') as f:
      f.write(TEMPLATE[0:i])
      f.write('["Path", "Parent", "Size", "Value"],\n')
      self._tree.write_report(f, None)
      f.write(TEMPLATE[i+len(marker):])

  @property
  def stats(self):
    return 'matched %d of %d lines' % (self._matches, self._lines)


TEMPLATE = r'''<!DOCTYPE html>
<title>Size of some binary</title>
<script src="https://www.gstatic.com/charts/loader.js"></script>
<style>
body {
  margin: 0;
}
#chart {
  width: 100vw;
  height: 100vh;
}
</style>
<div id="chart"></div>
<script>
google.charts.load('current', {'packages':['treemap']});
google.charts.setOnLoadCallback(draw_chart);
function draw_chart() {
  let data = google.visualization.arrayToDataTable([
    FIXME
  ]);
  let tree = new google.visualization.TreeMap(chart);
  tree.draw(data, {
    maxDepth: 2,
    showScale: false
  });
}
</script>
'''


def main():
  summarizer = Summarizer()
  with open('binary.txt') as in_file:
    for line in in_file:
      summarizer.process_line(line)
  summarizer.end()
  print(summarizer.stats)
  summarizer.write_report('index.html')


if __name__ == '__main__':
  main()
