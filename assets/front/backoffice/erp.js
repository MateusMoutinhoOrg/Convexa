// The behaviour of the ERP pages, served from the server itself because the
// Content-Security-Policy of the backoffice runs no script written inline.
// Every page works without it; it only hides what does not apply and fills
// what can be guessed.
(function () {
  // An element carrying data-show-when="field=value" shows only while the
  // select or input with id field holds value (several values separated by
  // "|"); its inputs are disabled while hidden, so they are not posted.
  document.querySelectorAll('[data-show-when]').forEach(function (element) {
    var rule = element.dataset.showWhen.split('=');
    var field = document.getElementById(rule[0]);
    if (!field) {
      return;
    }
    var values = rule[1].split('|');
    var sync = function () {
      var shown = field.type === 'checkbox' ? (field.checked ? 'on' : 'off') : field.value;
      var visible = values.indexOf(shown) !== -1;
      element.hidden = !visible;
      element.querySelectorAll('input, select, textarea').forEach(function (input) {
        input.disabled = !visible;
      });
    };
    field.addEventListener('change', sync);
    sync();
  });

  // A select carrying data-filter-by="field" offers only the options whose
  // data-group matches the value of the select or input with id field.
  document.querySelectorAll('select[data-filter-by]').forEach(function (select) {
    var field = document.getElementById(select.dataset.filterBy);
    if (!field) {
      return;
    }
    var sync = function () {
      Array.prototype.forEach.call(select.options, function (option) {
        if (!option.dataset.group) {
          return;
        }
        var fits = option.dataset.group === field.value;
        option.hidden = !fits;
        option.disabled = !fits;
        if (!fits && option.selected) {
          select.value = '';
        }
      });
    };
    field.addEventListener('change', sync);
    sync();
  });

  // A form carrying data-auto-submit sends itself when one of its selects
  // changes: the filters of a report.
  document.querySelectorAll('form[data-auto-submit] select').forEach(function (select) {
    select.addEventListener('change', function () {
      select.form.submit();
    });
  });
})();
