import std.array;
import std.stdio;

import csv;

class Project {
  this() {
    completion = new Task();
    tasks[completion.id] = completion;
  }

  void add(Task task) {
    if (task.id in tasks) {
      throw new Exception(
          std.string.format("task \"%s\" already exists", task.id));
    }
    completion.dependencyIds ~= task.id;
    tasks[task.id] = task;
  }

  void resolveDependencies() {
    foreach (task; tasks.byValue()) {
      task.dependencies = [];
      foreach (id; task.dependencyIds) {
        if (!(id in tasks)) {
            throw new Exception(
                std.string.format(
                    "no task with id \"%s\" found as predecessor of task %s",
                    id, task.id));
        }
        auto predecessor = tasks[id];
        if (std.algorithm.any!(existing => existing == predecessor)
                              (task.dependencies)) {
          throw new Exception(
              std.string.format(
                  "task %s listed multiple times as predecessor of task %s",
                  id, task.id));
        }
        task.dependencies ~= predecessor;
      }
    }
  }

  void checkCyclicDependencies() {
    foreach (task; tasks.byValue()) {
      Task[] work = [task];
      bool[Task] queued;

      while (work.length) {
        foreach (predecessor; work.front.dependencies) {
          if (predecessor == task) {
            throw new Exception(
                std.string.format(
                    "cyclic dependencies between task %s and task %s",
                    task.id, work.back.id));
          }

          if (predecessor in queued) {
            continue;
          }
          queued[predecessor] = true;
          work ~= predecessor;
        }

        work.popFront();
      }
    }
  }

  void checkDependencyCompletionConsistency() {
    foreach (task; tasks.byValue()) {
      if (task.status != Task.Status.Completed) {
        continue;
      }
      foreach (predecessor; task.dependencies) {
        if (predecessor.status != Task.Status.Completed) {
          throw new Exception(
              std.string.format(
                  "task %s is completed but predecessor task %s is not",
                  task.id, predecessor.id));
        }
      }
    }
  }

  Task completion;
  Task[string] tasks;
}

struct RACI {
  this(string descriptor) {
    this.descriptor = descriptor;
  }

  @property string responsible() {
    auto r = std.regex.regex(r"^RA?C?I?=([^,]*)");
    auto m = std.regex.match(descriptor, r);
    if (!m) {
      throw new Exception(
          std.string.format("could not find responsible person in \"%s\"",
                            descriptor));
    }
    return m.captures[1];
  }

  string descriptor;
}

class Task {
  this(string[] fields) {
    size_t n = 0;
    n++; // skip project
    this.id = fields[n++];
    this.dependencyIds =
        array(std.algorithm.filter!(s => s.length)
                                   (std.algorithm.splitter(fields[n++], ',')));
    this.timePlanned = std.conv.to!(float)(fields[n++]);

    string timeActual = fields[n++];
    this.timeActual = (timeActual.length == 0)
        ? 0.0f
        : std.conv.to!(float)(timeActual);
    this.raci = RACI(fields[n++]);
    this.who = fields[n++];
    this.startDate = fields[n++];
    this.endDate = fields[n++];
    this.title = fields[n++];
  }

  // Completion task
  this() {
    this.id = "ZZ";
    this.dependencyIds = [];
    this.timePlanned = 0;
    this.timeActual = 0;
    this.raci = RACI("");
    this.who = "";
    this.startDate = "";
    this.endDate = "";
    this.title = "COMPLETION";
  }

  // FIXME: Tasks are designed to only be hashed and compared within
  // the same project.
  override hash_t toHash() {
    return typeid(id).getHash(&id);
  }

  override bool opEquals(Object other) {
    return this is other;
  }

  override int opCmp(Object other) {
    return std.string.cmp(this.id, std.conv.to!Task(other).id);
  }

  @property bool isCompletion() {
    return id == "ZZ";
  }

  enum Status {
    NotStarted,
    Started,
    Completed
  }

  @property Status status() {
    if (endDate.length) {
      return Status.Completed;
    }
    if (startDate.length) {
      return Status.Started;
    }
    return Status.NotStarted;
  }

  @property float timeRemaining() {
      if (timeActual > timePlanned) {
        throw new Exception(
            std.string.format(
                "need to update estimated time for task %s", id));
      }
      return timePlanned - timeActual;
  }

  @property auto incompleteDependencies() {
    return (int delegate(ref Task) dg) {
      foreach (task; dependencies) {
        if (task.status == Status.Completed) {
          continue;
        }
        if (int result = dg(task)) {
          return result;
        }
      }
      return 0;
    };
  }

  @property auto incompleteDependenciesString() {
    return isCompletion
      ? ""
      : std.array.join(
            std.algorithm.map!(task => task.id)
                              (array(incompleteDependencies)), ",");
  }

  string id;
  string[] dependencyIds;
  Task[] dependencies;
  float timePlanned;
  float timeActual;
  RACI raci;
  string who;
  string startDate;
  string endDate;
  string title;
}

class Resources {
  this(string descriptor) {
    parse(descriptor);
  }

  string[][string] group;

  void debugPrint() {
    writeln("Resources");
    foreach (name, members; group) {
      writef("%s={ ", name);
      foreach (member; members) {
        writef("%s ", member);
      }
      writefln("}");
    }
  }

private:
  void parse(string descriptor) {
    auto r = std.regex.regex(r"(?P<group>[a-zA-Z]+)=(?P<members>[^;]+)", "g");
    auto s = std.regex.regex(r"[^,]+", "g");

    foreach (m; std.regex.match(descriptor, r)) {
      string[] members = [];
      foreach (n; std.regex.match(m.captures["members"], s)) {
        members ~= n.captures[0];
      }
      group[m.captures["group"]] = members;
    }
  }
}

class Schedule {
  this(Project project, Resources resources) {
    this.project = project;
    this.resources = resources;
  }

  class Entry {
    this(Task task) {
      this.task = task;
      this.end = 0.0f;
    }

    @property float start() {
      return end + task.timeRemaining;
    }

    Task task;
    float end;
  }

  void pushBackTo(Entry entry, float tMinus) {
    if (tMinus <= entry.end) {
      return;
    }
    entry.end = tMinus;
    foreach (predecessor; entry.task.incompleteDependencies) {
      if (!(predecessor in entries)) {
        continue;
      }
      pushBackTo(entries[predecessor], entry.start);
    }
  }

  void schedule() {
    // calculate path duration without resource leveling; schedule
    // late start with no buffers

    entries = null;

    void schedule(Task task, float tMinus) {
      Entry entry;
      if (task in entries) {
        entry = entries[task];
      } else {
        entry = entries[task] = new Entry(task);
      }

      pushBackTo(entry, tMinus);

      foreach (predecessor; task.incompleteDependencies) {
        schedule(predecessor, entry.start);
      }
    }
    schedule(project.completion, 0);

    // FIXME: do resource leveling
  }

  @property Task[] criticalPath() {
    Task task = project.completion;
    pathDuration(task); // populate criticalPathLink
    Task[] path = [];
    while (task) {
      path = task ~ path;
      task = criticalPathLink[task];
    }
    return path;
  }

  void debugPrint() {
    writeln("Critical path:");
    foreach (task; criticalPath) {
      writefln("%2s %s", task.id, task.title);
    }
    writeln();

    writeln("Schedule:");
    writeln("Start End   Preds. Task");
    writeln("----- ----- ------ -----");
    foreach (entry; entries) {
      writefln("%5.1f %5.1f %6s %2s %-58.58s",
               entry.start, entry.end, entry.task.incompleteDependenciesString,
               entry.task.id, entry.task.title);
    }
  }

  float pathDuration(Task task) {
    if (task in taskPathDuration) {
      return taskPathDuration[task];
    }

    float maximumPredecessorTime = 0.0f;
    Task maximumPredecessor = null;
    foreach (predecessor; task.incompleteDependencies) {
      float predecessorTime = pathDuration(predecessor);
      if (predecessorTime > maximumPredecessorTime) {
        maximumPredecessorTime = predecessorTime;
        maximumPredecessor = predecessor;
      }
    }
    criticalPathLink[task] = maximumPredecessor;

    final switch (task.status) {
    case Task.Status.NotStarted:
    case Task.Status.Started:
      return taskPathDuration[task] =
          task.timeRemaining + maximumPredecessorTime;

    case Task.Status.Completed:
      return taskPathDuration[task] = 0.0f;
    }
  }

  Project project;
  Resources resources;
  float[Task] taskPathDuration; // critical path, no resource leveling
  Task[Task] criticalPathLink;  // critical path, no resource leveling
  Entry[Task] entries;
}

void main() {
  string[][] fields = CSVParser(std.stdio.stdin).toArray();

  // parse resources descriptor
  auto resources = new Resources(fields[1][0]);

  // strip header lines
  fields = fields[2 .. $];

  if (!fields.length) {
    writefln("No tasks.");
    return;
  }

  // check all tasks are for the same project
  string projectId = fields[0][0];
  foreach (row; fields[1 .. $]) {
    if (row[0] != projectId) {
      writefln("Can't schedule multiple projects \"%s\" and \"%s\".",
               projectId, row[0]);
      return;
    }
  }

  // build tasks
  auto tasks = array(std.algorithm.map!(x => new Task(x))(fields));

  auto project = new Project();
  foreach (task; tasks) {
    project.add(task);
  }

  project.resolveDependencies();
  project.checkCyclicDependencies();
  project.checkDependencyCompletionConsistency();

  auto schedule = new Schedule(project, resources);

  // print summary statistics

  resources.debugPrint();
  writeln();

  schedule.schedule();
  schedule.debugPrint();
  writeln();

  struct PerPersonStats {
    uint numTasks;
    float time = 0.0f;
  }
  PerPersonStats[string] perPersonStats;
  foreach (task; tasks) {
    if (!(task.raci.responsible in perPersonStats)) {
      perPersonStats[task.raci.responsible] = PerPersonStats.init;
    }
    perPersonStats[task.raci.responsible].numTasks++;
    perPersonStats[task.raci.responsible].time += task.timePlanned;
  }
  writefln("Person     # Tasks Time (h)");
  writefln("---------- ------- --------");
  foreach (person, stats; perPersonStats) {
    writefln("%-10s %7d %8.1f", person, stats.numTasks, stats.time);
  }
}
