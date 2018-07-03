'use strict';
(function() {
  const start_checked = [
    'core2', 'td', 'dad', 'cac', 'hap', 'dtwn', '23s', 'cd'
  ];

  var sessionId = null;
  var sideId = null;
  var faction = null;
  var identity = null;
  var remaining_influence = null;

  const neutralRegex = /neutral/i;

  const sideCorp = 'corp';
  const sideRunner = 'runner';

  const hidden_class = 'hidden';
  const loading_div = document.getElementById('loading');
  const packs_div = document.getElementById('choose-packs');
  const packs_form = document.getElementById('packs-form');
  const packs_field = document.getElementById('packs-field');
  const sides_div = document.getElementById('choose-side');
  const factions_div = document.getElementById('choose-faction');
  const factions_form = document.getElementById('factions-form');
  const factions_field = document.getElementById('faction-field');
  const identities_div = document.getElementById('choose-identity');
  const identities_form = document.getElementById('identities-form');
  const identities_field = document.getElementById('identity-field');
  const draft_div = document.getElementById('draft-cards');
  const draft_form = document.getElementById('cards-form');
  const draft_field = document.getElementById('cards-field');
  
  const corp_link_id = 'corp-link';
  const corp_link = document.getElementById(corp_link_id);
  const runner_link_id = 'runner-link';
  const runner_link = document.getElementById(runner_link_id);
  
  function start() {
    // Set up the packs choice form
    var oReq = new XMLHttpRequest();
    oReq.addEventListener('load', load_packs(oReq));
    oReq.open('GET', 'data/packs.json');
    oReq.send();
    packs_form.addEventListener('submit', choose_packs);
    factions_form.addEventListener('submit', choose_faction);
    identities_form.addEventListener('submit', choose_identity);
    [corp_link, runner_link].forEach(function(link) {
      link.addEventListener('click', choose_side);
    });
  }

  function load_packs(xhr) {
    return function() {
      if (xhr.readyState !== 4 && xhr.status !== 200) {
        return;
      }
      const json = JSON.parse(xhr.responseText);
      json.forEach(function(pack) {
        const entry = makeInput('have-packs');
        entry.setAttribute('value', pack.Code);
        if (start_checked.includes(pack.Code)) {
          entry.setAttribute('checked', true);
        }
        packs_field.appendChild(entry);
        packs_field.appendChild(document.createTextNode(pack.Name));
        packs_field.appendChild(document.createElement('br'));
      });
      removeClass(packs_div, hidden_class);
    };
  }

  function choose_packs(event) {
    const packs = event.target.querySelectorAll("input[type='checkbox']:checked");
    const url_begin = 'draft/withPacks/';
    const url_end = '/factions';

    packs_div.className += hidden_class;
    removeClass(sides_div, hidden_class);

    var url = url_begin;
    var first = true;
    packs.forEach(function(pack) {
      if (first) {
        first = false;
      } else {
        url += ',';
      }
      url += pack.value;
    });
    url += url_end;
    console.log('url: ', url);
    
    var oReq = new XMLHttpRequest();
    oReq.addEventListener('load', load_factions(oReq));
    oReq.open('GET', url);
    oReq.send();

    // Don't change the URL
    event.preventDefault();
    return false;
  }

  function choose_side(event) {
    // Don't change the URL
    event.preventDefault();
    if (event.target.id == corp_link_id) {
      sideId = sideCorp;
    } else if (event.target.id == runner_link_id) {
      sideId = sideRunner;
    } else {
      console.error('Unexpected event: ', event);
      return false;
    }

    sides_div.className += hidden_class;
    removeClass(loading_div, hidden_class);

    // TODO: filter identities

    return false;
  }

  function load_factions(xhr) {
    return function() {
      if (xhr.readyState !== 4 && xhr.status !== 200) {
        return;
      }
      const json = JSON.parse(xhr.responseText);
      sessionId = json.Session;
      json.Factions.filter(function(faction) {
        return sideId == null ||
          (faction.side_code == sideId &&
           faction.code.startsWith("neutral-") == false);
      }).forEach(function(faction) {
        const entry = makeRadio('faction-choice');
        entry.setAttribute('value', JSON.stringify(faction));
        factions_field.appendChild(entry);
        factions_field.appendChild(document.createTextNode(faction.name));
        factions_field.appendChild(document.createElement('br'));
      });
      loading_div.className += hidden_class;
      removeClass(factions_div, hidden_class);
    };
  }

  function choose_faction(event) {
    event.preventDefault();
    if (sessionId === null) {
      return false;
    }
    const factions_chosen = event.target.querySelectorAll("input[type='radio']:checked");
    if (factions_chosen.length < 1) {
      return false;
    }
    
    factions_div.className += hidden_class;
    
    faction = JSON.parse(factions_chosen[0].value);
    console.dir(faction);
    const url = 'draft/session/' + sessionId + '/faction/' + faction.code + '/identities';
    console.log('choose_faction url: ', url);
    
    var oReq = new XMLHttpRequest();
    oReq.addEventListener('load', load_identities(oReq));
    oReq.open('GET', url);
    oReq.send();
    
    return false;
  }

  function load_identities(xhr) {
    return function() {
      if (xhr.readyState !== 4 && xhr.status !== 200) {
        return;
      }
      const json = JSON.parse(xhr.responseText);
      console.dir('identities:', json);
      json.Identities.forEach(function(identity) {
        const entry = makeRadio('identity-choice');
        entry.setAttribute('value', JSON.stringify(identity));
        identities_field.appendChild(entry);
        identities_field.appendChild(document.createTextNode(identity.Title));
        identities_field.appendChild(document.createElement('br'));
        const img = document.createElement('img');
        if ('AltImageUrl' in identity) {
          img.setAttribute('src', identity.AltImageUrl);
        } else {
          img.setAttribute('src', identity.ImageUrl);
        }
        img.setAttribute('alt', identity.Text);
        identities_field.appendChild(img);
        identities_field.appendChild(document.createElement('br'));
      });
      removeClass(identities_div, hidden_class);
    };
  }

  function choose_identity(event) {
    event.preventDefault();
    if (sessionId === null) {
      return false;
    }
    
    identities_div.className += hidden_class;

    const identities_chosen = event.target.querySelectorAll("input[type='radio']:checked");
    if (identities_chosen.length < 1) {
      return false;
    }
    identity = JSON.parse(identities_chosen[0].value);
    console.dir(identity);
    remaining_influence = parseInt(identity.Details.influence_limit, 10);

    const url = 'draft/session/' + sessionId + '/faction/' + faction.code + '/withInfluence/' + remaining_influence;
    
    var oReq = new XMLHttpRequest();
    oReq.addEventListener('load', load_cards(oReq));
    oReq.open('GET', url);
    oReq.send('[]');

    return false;
  }

  function load_cards(xhr) {
    return function() {
      if (xhr.readyState !== 4 && xhr.status !== 200) {
        return;
      }
      const json = JSON.parse(xhr.responseText);
      console.dir('cards:', json);
      json.Cards.forEach(function(card, i) {
        const entry = makeRadio('card-choice');
        entry.setAttribute('value', JSON.stringify(json.CardCodeQuantities[i]));
        draft_field.appendChild(entry);
        draft_field.appendChild(document.createTextNode(card.Title));
        draft_field.appendChild(document.createElement('br'));
        if ('AltImageUrl' in card) {
          const img = document.createElement('img');
          img.setAttribute('src', card.AltImageUrl);
          img.setAttribute('alt', card.Text);
          draft_field.appendChild(img);
          draft_field.appendChild(document.createElement('br'));
        }
        if ('ImageUrl' in card) {
          const img = document.createElement('img');
          img.setAttribute('src', card.ImageUrl);
          img.setAttribute('alt', card.Text);
          draft_field.appendChild(img);
        }
        draft_field.appendChild(document.createElement('br'));
      });
      removeClass(draft_div, hidden_class);
    };
  }

  function choose_card() {
  }
  
  function removeClass(elements, myClass) {
    // if there are no elements, we're done
    if (!elements) { return; }

    if (typeof(elements) === 'string') {
      // if we have a selector, get the chosen elements
      elements = document.querySelectorAll(elements);
    } else if (elements.tagName) {
      // if we have a single DOM element, make it an array to simplify behavior
      elements=[elements];
    }

    // create pattern to find class name
    var reg = new RegExp('(^| )'+myClass+'($| )','g');

    // remove class from all chosen elements
    for (var i=0; i<elements.length; i++) {
      elements[i].className = elements[i].className.replace(reg,' ');
    }
  }

  function makeRadio(name) { return makeInput(name, 'radio'); }

  function makeInput(name, type) {
    if (type === undefined) {
      type = 'checkbox';
    }
    const entry = document.createElement('input');
    entry.setAttribute('type', type);
    entry.setAttribute('name', name);
    return entry;
  }

  start();
})();
