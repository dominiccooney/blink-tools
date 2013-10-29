import std.array;
import std.stdio;

import csv;

class Project {
  void add(Task task) {
    if (task.id in tasks) {
      throw new Exception(
          std.string.format("task \"%s\" already exists", task.id));
    }
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
    this.project = fields[n++];
    this.id = fields[n++];
    this.dependencyIds =
        array(std.algorithm.filter!(s => s.length)
                                   (std.algorithm.splitter(fields[n++], ',')));
    this.time = std.conv.to!(float)(fields[n++]);
    n++; // skip time--actual (hours)
    this.raci = RACI(fields[n++]);
    this.who = fields[n++];
    n++; // skip start date
    n++; // skip end date
    this.title = fields[n++];
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

  string project;
  string id;
  string[] dependencyIds;
  Task[] dependencies;
  float time;
  RACI raci;
  string who;
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

  void schedule() {
    // calculate path duration without resource leveling
    foreach (task; project.tasks.byValue()) {
      pathDuration(task);
    }
  }

  @property Task[] criticalPath() {
    float maxTime = 0;
    Task maxTask = null;

    foreach (task; project.tasks.byValue()) {
      float duration = pathDuration(task);
      if (duration > maxTime) {
        maxTime = duration;
        maxTask = task;
      }
    }

    Task[] path = [];
    while (maxTask) {
      path = maxTask ~ path;
      maxTask = criticalPathLink[maxTask];
    }

    return path;
  }

  void debugPrint() {
    writeln("Critical path:");
    foreach (task; criticalPath) {
      writefln("%2s %s", task.id, task.title);
    }
  }

  float pathDuration(Task task) {
    if (task in taskPathDuration) {
      return taskPathDuration[task];
    }

    float maximumPredecessorTime = 0.0f;
    Task maximumPredecessor = null;
    foreach (predecessor; task.dependencies) {
      float predecessorTime = pathDuration(predecessor);
      if (predecessorTime > maximumPredecessorTime) {
        maximumPredecessorTime = predecessorTime;
        maximumPredecessor = predecessor;
      }
    }
    criticalPathLink[task] = maximumPredecessor;
    return taskPathDuration[task] = task.time + maximumPredecessorTime;
  }

  Project project;
  Resources resources;
  float[Task] taskPathDuration; // critical path, no resource leveling
  Task[Task] criticalPathLink;  // critical path, no resource leveling
}

void main() {
  string[][] fields = CSVParser(std.stdio.stdin).toArray();

  // parse resources descriptor
  auto resources = new Resources(fields[1][0]);

  // strip header lines
  fields = fields[2 .. $];

  // build tasks
  auto tasks = array(std.algorithm.map!(x => new Task(x))(fields));

  if (!tasks.length) {
    writefln("No tasks.");
    return;
  }

  // check all tasks are for the same project
  string projectId = tasks[0].project;
  foreach (task; tasks[1 .. $]) {
    if (task.project != projectId) {
      writefln("Can't schedule multiple projects \"%s\" and \"%s\".",
               projectId, task.project);
      return;
    }
  }

  auto project = new Project();
  foreach (task; tasks) {
    project.add(task);
  }

  project.resolveDependencies();
  project.checkCyclicDependencies();

  auto schedule = new Schedule(project, resources);

  // print summary statistics

  resources.debugPrint();
  writeln();

  writefln("%d task(s)", fields.length);
  writeln();

  schedule.debugPrint();

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
    perPersonStats[task.raci.responsible].time += task.time;
  }
  writefln("Person     # Tasks Time (h)");
  writefln("---------- ------- --------");
  foreach (person, stats; perPersonStats) {
    writefln("%-10s %7d %8.1f", person, stats.numTasks, stats.time);
  }
}
