'use strict';

const body = document.body;
body.addEventListener('dragenter', (e) => {
  e.stopPropagation();
  e.preventDefault();
}, false);
body.addEventListener('dragleave', (e) => {
  console.log(e);
}, false);
body.addEventListener('dragover', (e) => {
  e.stopPropagation();
  e.preventDefault();
}, false);
body.addEventListener('drop', handleDrop, false);

function handleDrop(e) {
  e.preventDefault();
  for (let file of e.dataTransfer.files) {
    load(file);
    return;
  }
}

async function load(file) {
  document.title = file.name;
  let lines = (await file.text()).split('\n');
  const pattern = /^(?<datetime>\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{3})\s+(?<pid>\d+)\s+(?<tid>\d+)\s/;
  logLinesView.textContent = '';
  for (const line of lines) {
    const logLine = createLogLineView(line);
    logLinesView.appendChild(logLine);

    const match = line.match(pattern);
    if (!match) {
      continue;
    }
    logLine.classList.add('valid');
    logLine.date = new Date('2021-' + match.groups['datetime']);
  }
}

function* logLineViews() {
  for (let line = logLinesView.firstElementChild; line; line = line.nextElementSibling) {
    yield line;
  }
}

logLinesView.addEventListener('click', handleLogLineClick);
function handleLogLineClick(event) {
  const target = event.target;
  if (!(target.classList.contains('logLine') && target.classList.contains('valid'))) {
    return;
  }
  for (let line of logLineViews()) {
    displayRelativeTo(line, target.date);
  }
}

function createLogLineView(text) {
  const view = document.createElement('div');
  view.text = text;
  view.classList.add('logLine');
  const delta = document.createElement('div');
  delta.classList.add('delta');
  view.append(delta);
  view.append(text);
  return view;
}

function displayRelativeTo(line, date) {
  if (!line.classList.contains('valid')) {
    return;
  }
  const delta_msec = line.date - date;
  const delta_element = line.querySelector('.delta');
  if (delta_msec == 0) {
    delta_element.classList.remove('show');
    return;
  }
  if (delta_msec < 0) {
    delta_element.classList.add('neg');
  } else {
    delta_element.classList.remove('neg');
  }
  delta_element.textContent = delta_msec;
  delta_element.classList.add('show');
}

search.addEventListener('input', searchChanged, false);
function searchChanged(event) {
  if (search.value === '') {
    for (let line of logLineViews()) {
      line.classList.remove('hide');
    }
  }
  let pattern;
  try {
    pattern = new RegExp(search.value);
  } catch (e) {
    search.classList.add('invalid');
    return;
  }
  search.classList.remove('invalid');
  for (let line of logLineViews()) {
    if (pattern.test(line.text)) {
      line.classList.remove('hide');
    } else {
      line.classList.add('hide');
    }
  }
}
