{% set update_interval = "10" %}
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8" />
  <meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0">

  <title>{% block title %}RuneScape Signatures{% endblock %}</title>
  {% block default_stylesheets %}
  <link rel="stylesheet" type="text/css" href="/assets/css/semantic.min.css">
  <link rel="stylesheet" type="text/css" href="/assets/css/style.css">
  {% endblock %}
  {% block stylesheets %}{% endblock %}
</head>
<body>
  <div class="ui main text container">
    <h1 class="ui header">RuneScape Signatures</h1>
    <p>Create automatically updated signatures to show off your RuneScape skill goals!</p>
    <p>All signatures are currently updated every <strong>{{ update_interval }}</strong> minutes.</p>
  </div>
  {% block body %}{% endblock %}
  <a href="https://github.com/cubeee/go-sig"><img style="position: absolute; top: 0; right: 0; border: 0;" src="https://camo.githubusercontent.com/a6677b08c955af8400f44c6298f40e7d19cc5b2d/68747470733a2f2f73332e616d617a6f6e6177732e636f6d2f6769746875622f726962626f6e732f666f726b6d655f72696768745f677261795f3664366436642e706e67" alt="Fork me on GitHub" data-canonical-src="https://s3.amazonaws.com/github/ribbons/forkme_right_gray_6d6d6d.png"></a>
</body>
</html>
