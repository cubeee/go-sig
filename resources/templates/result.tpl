{% extends "_base.tpl" %}

{% block title %}Results :: RuneScape Signatures{% endblock %}
{% block body %}

<div class="ui container">
  <div class="ui one column centered grid">
    <div class="twelve wide column">
      <div class="ui raised segment two column centered internally celled grid">
        <div class="eight wide column">
          <h3>Result</h3>
          <form class="ui large form">
            <div class="field">
              <div class="ui fluid labeled small input">
                <div class="ui label">Direct link:</div>
                <input type="text" value="{{ base_url }}{{ url }}">
              </div>
            </div>
            <div class="field">
              <div class="ui fluid labeled small input">
                <div class="ui label">BBCode:</div>
                <input type="text" value="[img]{{ base_url }}{{ url }}[/img]">
              </div>
            </div>
            <div class="field">
              <div class="ui fluid labeled small input">
                <div class="ui label">BBCode with URL:</div>
                <input type="text" value='[url="{{ base_url }}"][img]{{ base_url }}{{ url }}[/img][/url]'>
              </div>
            </div>
            <div class="field">
              <a href="/" class="ui primary button">Back to home</a>
            </div>
          </form>
        </div>
        <div class="eight wide column">
          <img class="ui centered image" src="{{ base_url }}{{ url }}" alt="Example" />
        </div>
      </div>
    </div>
  </div>
</div>

<a href="https://github.com/cubeee/go-sig"><img style="position: absolute; top: 0; right: 0; border: 0;" src="https://camo.githubusercontent.com/a6677b08c955af8400f44c6298f40e7d19cc5b2d/68747470733a2f2f73332e616d617a6f6e6177732e636f6d2f6769746875622f726962626f6e732f666f726b6d655f72696768745f677261795f3664366436642e706e67" alt="Fork me on GitHub" data-canonical-src="https://s3.amazonaws.com/github/ribbons/forkme_right_gray_6d6d6d.png"></a>

{% endblock %}
