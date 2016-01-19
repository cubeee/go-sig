var xpGoalInputMask = function() {
  $(".tooltip-sig-xp-goal").inputmask({
    mask: "9{1,3}[,9{0,3}][,9{0,3}]",
    placeholder: "",
    greedy: false
  });
}

var levelGoalInputMask = function() {
  $(".tooltip-sig-level-goal").inputmask({
    mask: "9{1,3}",
    placeholder: "",
    greedy: false
  });
}

$(document).ready(function() {
  $(".ui.dropdown").dropdown();

  xpGoalInputMask();
  levelGoalInputMask();

  $("#username-field").inputmask('Regex', {
    regex: "[a-zA-Z0-9-_ ]{1,12}",
    placeholder: "",
    greedy: false
  });

  $("#goal-dropdown > .item").click(function(e) {
    var goalField = $('#tooltip-skill-goal');
    var newItem = e.target.outerText;
    var goalClass;
    var refreshFunction = null;
    if (newItem == 'Experience') {
      goalClass = 'tooltip-sig-xp-goal';
      refreshFunction = xpGoalInputMask;
    } else if (newItem == 'Level') {
      goalClass = 'tooltip-sig-level-goal';
      refreshFunction = levelGoalInputMask;
    }
    goalField.attr('class', goalClass);
    goalField.val('');
    if (refreshFunction != null) {
      refreshFunction.call();
    }
  });
});
