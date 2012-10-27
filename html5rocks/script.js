[].forEach.call(
    document.querySelectorAll('.pattern-bg'),
    function (element) {
      element.classList.remove('pattern-bg');
      element.classList.add('new-pattern-bg');
    });

[].forEach.call(
    document.querySelectorAll('.pattern-bg-lighter'),
    function (element) {
      element.classList.remove('pattern-bg-lighter');
      element.classList.add('new-pattern-bg-lighter');
    });
