'use strict';
(function() {
  var sessionId = null;
  var sideId = null;
  var faction = null;
  var identity = null;
  var remaining_influence = null;
  const drafted_cards = new Set();
  const drafted_deck = new Map();
  var drafted_count = 0;
  var agenda_points = 0;

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
  const draft_count = document.getElementById('draft-count');
  const agenda_count = document.getElementById('agenda-count');
  const draft_influence = document.getElementById('influence-count');
  const draft_div = document.getElementById('draft-cards');
  const draft_form = document.getElementById('cards-form');
  const draft_field = document.getElementById('cards-field');
  const deck_div = document.getElementById('card-deck');
  const deck_text = document.getElementById('deck-export');
  
  const corp_link_id = 'corp-link';
  const corp_link = document.getElementById(corp_link_id);
  const runner_link_id = 'runner-link';
  const runner_link = document.getElementById(runner_link_id);
  
  function start() {
    document.getElementById('choose-packs-uncheck').addEventListener('click', function(event) {
      event.preventDefault();
      event.target.parentNode.querySelectorAll("input[type='checkbox']:checked").forEach(function(input) {
        input.checked = false;
      });
      return false;
    });
    packs_form.addEventListener('submit', choose_packs);
    factions_form.addEventListener('submit', choose_faction);
    identities_form.addEventListener('submit', choose_identity);
    draft_form.addEventListener('submit', choose_card);
    [corp_link, runner_link].forEach(function(link) {
      link.addEventListener('click', choose_side);
    });
  }

  function choose_packs(event) {
    const packs = event.target.querySelectorAll("input[type='checkbox']:checked");
    const url_begin = 'draft/withPacks/';
    const url_end = '/sides';

    packs_div.className += hidden_class;
    removeClass(loading_div, hidden_class);

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
    
    var oReq = new XMLHttpRequest();
    oReq.addEventListener('load', load_sides(oReq));
    oReq.open('GET', url);
    oReq.send();

    // Don't change the URL
    event.preventDefault();
    return false;
  }

  function load_sides(xhr) {
    return function() {
      if (xhr.readyState !== 4 && xhr.status !== 200) {
        return;
      }
      const json = JSON.parse(xhr.responseText);
      sessionId = json.Session;
      loading_div.className += hidden_class;
      removeClass(sides_div, hidden_class);
    };
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
    if (sideId == sideCorp) {
      removeClass(document.getElementById('corp-stats'), hidden_class);
    }

    const url = `draft/session/${sessionId}/side/${sideId}/factions`;

    var oReq = new XMLHttpRequest();
    oReq.addEventListener('load', load_factions(oReq));
    oReq.open('GET', url);
    oReq.send();

    return false;
  }

  function load_factions(xhr) {
    return function() {
      if (xhr.readyState !== 4 && xhr.status !== 200) {
        return;
      }
      const json = JSON.parse(xhr.responseText);
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
    
    const url = `draft/session/${sessionId}/side/${sideId}/faction/${faction.code}/identities`;
    
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
      json.Identities.forEach(function(identity) {
        const entry = makeRadio('identity-choice');
        entry.setAttribute('value', JSON.stringify(identity));
        identities_field.appendChild(entry);
        identities_field.appendChild(document.createTextNode(identity.Title));
        identities_field.appendChild(document.createElement('br'));
        const img = document.createElement('img');
        img.setAttribute('alt', identity.Text);
        if ('ImageUrl' in identity && identity.ImageUrl !== '') {
          img.setAttribute('src', identity.ImageUrl);
          if ('AltImageUrl' in identity && identity.AltImageUrl !== '') {
            img.addEventListener('error', function(event) {
              event.target.setAttribute('src', identity.AltImageUrl);
            });
          }
        } else if ('AltImageUrl' in identity && identity.AltImageUrl !== '') {
          img.setAttribute('src', identity.AltImageUrl);
        } else {
          console.error('No image url to set for identity', identity);
        }
        identities_field.appendChild(img);
        identities_field.appendChild(document.createElement('br'));
      });
      removeClass(identities_div, hidden_class);
    };
  }

  function update_influence() {
    draft_influence.innerHTML = `${remaining_influence}/${identity.Details.influence_limit}`;
  }

  function update_count() {
    draft_count.innerHTML = `${drafted_count}/${identity.Details.minimum_deck_size}`;
  }

  function update_agendas() {
    const deck_size = Math.max(drafted_count, identity.Details.minimum_deck_size);
    const req_points = required_agenda_points(deck_size);
    agenda_count.innerHTML = `${agenda_points}/${req_points.min}`;
  }

  function required_agenda_points(deck_size) {
    if (40 <= deck_size && deck_size <= 44) {
      return {min: 18, max: 19};
    } else if (45 <= deck_size && deck_size <= 49) {
      return {min: 20, max: 21};
    } else if (50 <= deck_size && deck_size <= 54) {
      return {min: 22, max: 23};
    } else if (deck_size >= 55) {
      const extra = Math.floor((deck_size - 55) / 5);
      return {min: 22 + extra, max: 23 + extra};
    } else {
      console.error(`Invalid deck size: ${deck_size}`);
      return {min: 18, max: 19};
    }
  }

  function choose_identity(event) {
    event.preventDefault();
    if (sessionId === null) {
      return false;
    }
    
    const identities_chosen = event.target.querySelectorAll("input[type='radio']:checked");
    if (identities_chosen.length < 1) {
      return false;
    }
    identities_div.className += hidden_class;
    identity = JSON.parse(identities_chosen[0].value);
    remaining_influence = parseInt(identity.Details.influence_limit, 10);
    deck_text.value += `${identity.Title}\n`;

    const url = `draft/session/${sessionId}/side/${sideId}/faction/${faction.code}/withInfluence/${remaining_influence}/cards`;

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
      json.Cards.forEach(function(card, i) {
        const entry = makeRadio('card-choice');
        entry.setAttribute('data', JSON.stringify(card));
        entry.setAttribute('value', JSON.stringify(json.CardCodeQuantities[i]));
        draft_field.appendChild(entry);
        draft_field.appendChild(document.createTextNode(card.Title));
        draft_field.appendChild(document.createElement('br'));
        draft_field.appendChild(card_img(card));
        draft_field.appendChild(document.createElement('br'));
      });
      update_influence();
      update_count();
      if (sideId == sideCorp) {
        update_agendas();
      }
      removeClass(draft_div, hidden_class);
    };
  }

  function card_img(card) {
    const img = document.createElement('img');
    img.setAttribute('alt', card.Text);
    if ('ImageUrl' in card && card.ImageUrl !== '') {
      img.setAttribute('src', card.ImageUrl);
      if ('AltImageUrl' in card && card.AltImageUrl !== '') {
        img.addEventListener('error', function(event) {
          event.target.setAttribute('src', card.AltImageUrl);
        });
      }
    } else if ('AltImageUrl' in card && card.AltImageUrl !== '') {
      img.setAttribute('src', card.AltImageUrl);
    } else {
      console.error('No image url to set for card', card);
    }
    return img;
  }

  function choose_card(event) {
    event.preventDefault();
    if (sessionId === null) {
      return false;
    }
    
    const card_chosen = event.target.querySelectorAll("input[type='radio']:checked");
    if (card_chosen.length < 1) {
      return false;
    }
    draft_div.className += hidden_class;
    removeClass(deck_div, hidden_class);

    const all_cards = event.target.querySelectorAll("input[type='radio']");
    const retCards = [];
    all_cards.forEach(function(card_elmnt) {
      const ccq = JSON.parse(card_elmnt.value);
      const card = JSON.parse(card_elmnt.attributes.getNamedItem('data').value);
      if (card_elmnt.checked) {
        add_drafted_card(card, ccq);
        ccq.Quantity -= 1;
        if (ccq.Quantity > 0) {
          retCards.push(ccq);
        }
        if (ccq.Faction !== faction.code) {
          remaining_influence -= parseInt(card.Details.faction_cost, 10);
        }
        if (card.Types.includes('agenda')) {
          agenda_points += parseInt(card.Details.agenda_points, 10);
        }
      } else {
        retCards.push(ccq);
      }
    });

    while (draft_field.firstChild) {
      draft_field.firstChild.remove();
    }

    const url = `draft/session/${sessionId}/side/${sideId}/faction/${faction.code}/withInfluence/${remaining_influence}/cards`;

    var oReq = new XMLHttpRequest();
    oReq.addEventListener('load', load_cards(oReq));
    oReq.open('POST', url);
    oReq.send(JSON.stringify(retCards));

    return false;
  }

  function add_drafted_card(card, ccq) {
    drafted_count += 1;
    drafted_cards.add(card);
    if (drafted_deck.has(card.Code)) {
      const quantity = drafted_deck.get(card.Code);
      drafted_deck.set(card.Code, quantity + 1);
      var new_text = '';
      deck_text.value.split('\n').forEach(function(line) {
        if (line.length === 0) return;
        if (line.endsWith(card.Title)) {
          const new_count = 1 + parseInt(line[0], 10);
          new_text += `${new_count}x ${card.Title}\n`;
        } else {
          new_text += line + '\n';
        }
      });
      deck_text.value = new_text;
    } else {
      drafted_deck.set(card.Code, 1);
      deck_div.appendChild(card_img(card));
      deck_div.appendChild(document.createElement('br'));
      deck_text.value += `1x ${card.Title}\n`;
    }
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
