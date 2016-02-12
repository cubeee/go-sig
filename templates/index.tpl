{% extends "_base.tpl" %}

{% block title %}RuneScape Signatures{% endblock %}
{% block body %}

{% macro skill_dropdown(skills) %}
<div class="ui selection dropdown" tabindex="0" style="width: 100%; border-radius: 0;">
  <select name="skill">
    {% for skill in skills %}
    <option value="{{ skill|lower }}">{{ skill }}</option>
    {% endfor %}
  </select>
  <i class="dropdown icon"></i>
  <div class="text">Attack</div>
  <div class="menu transition hidden" tabindex="0">
    {% for skill in skills %}
    <div class="item" data-value="{{ skill|lower }}">{{ skill }}</div>
    {% endfor %}
  </div>
</div>
{% endmacro %}
<div class="ui container">
  <div class="ui two column grid">
    <!-- Tooltip goal -->
    <div class="column">
      <div class="ui raised segment three column grid">
        <div class="ten wide column">
          <h3>Skill interface tooltip</h3>
          <form id="generator" class="ui large form" action="/tooltip/create" method="POST">
            <div class="field">
              <div class="ui fluid labeled small input">
                <div class="ui label">Username:</div>
                <input id="username-field" type="text" name="username">
              </div>
            </div>
            <div class="field">
              <div class="ui fluid labeled small input">
                <div class="ui label">Skill:</div>
                {{ skill_dropdown(skills) }}
              </div>
            </div>
            <div class="field">
              <div class="ui fluid labeled small input">
                <div class="ui label">Goal:</div>
                <input id="tooltip-skill-goal" class="tooltip-sig-level-goal" type="text" name="goal" style="border-radius: 0;">
                <div class="ui dropdown label" tabindex="0" style="border-radius: 0;">
                  <div class="text">Level</div>
                  <i class="dropdown icon"></i>
                  <div id="goal-dropdown" class="menu transition hidden" tabindex="-1">
                    <div class="item active selected">Level</div>
                    <div class="item">Experience</div>
                  </div>
                </div>
              </div>
            </div>
            {% if has_aes %}
            <div class="field">
              <div class="ui checkbox">
                <input type="checkbox" name="hide">
                <label>Hide username</label>
              </div>
            </div>
            {% endif %}
            <div class="field">
              <input type="submit" name="submit" class="ui button" value="Create">
            </div>
          </form>
        </div>
        <div class="right floated six wide column">
          <img class="ui top aligned image" src="/assets/img/box_example.png" alt="Example" />
        </div>
      </div>
    </div>
    <!-- Multiple goals -->
    <div class="column">
      <div class="ui raised segment one column grid">
        <div class="column">
          <h3>Multiple skill goals in one - coming soon!</h3>
        </div>
      </div>
    </div>
  </div>
</div>

<script type="text/javascript" src="//cdnjs.cloudflare.com/ajax/libs/jquery/2.2.0/jquery.js"></script>
<script type="text/javascript" src="//cdnjs.cloudflare.com/ajax/libs/jquery.inputmask/3.2.5/jquery.inputmask.bundle.min.js"></script>
<script type="text/javascript" src="/assets/js/semantic.min.js"></script>
<script type="text/javascript" src="/assets/js/script.js"></script>
{% endblock %}
