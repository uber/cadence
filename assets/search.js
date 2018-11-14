---
---

var docs = [
  {% assign doc_pages = site.html_pages | where_exp: "page", "page.url contains '/docs'" %}
  {% for page in doc_pages %}
  {
    id: {{ forloop.index0 }},
    title: "{{ page.title | xml_escape }}",
    content: "{{ page.content | markdownify | strip_html | xml_escape | newline_to_br | strip_newlines | replace: '<br />', ' ' }}",
    url: "{{ page.url | absolute_url | xml_escape }}",
    relUrl: "{{ page.url | xml_escape }}"
  }{% if forloop.last %}{% else %},
  {% endif %}{% endfor %}
]

var index = lunr(function() {
  this.ref('id')
  this.field('title', { boost: 20 })
  this.field('content', { boost: 10 })
  this.field('url')

  for (let d of docs) {
    this.add(d)
  }
});

(function() {
  var searchContainer = document.querySelector('.doc-search')
      resultsEl = searchContainer.querySelector('ul.search-results'),
      input = searchContainer.querySelector('input')

  function handler() {
    if (!input.value) {
      resultsEl.innerHTML = ''
      return
    }

    var results = index.search(input.value)
    if (!results.length) {
      // try a simple substring match on the title if lunr didn't find interesting matches
      let lcVal = input.value.toLowerCase()
      results = docs.filter(d => d.title.toLowerCase().includes(lcVal)).map((_, ref) => ({ ref }))
    }

    resultsEl.setAttribute('data-result-count', String(results.length))
    resultsEl.innerHTML = results.length ? results.map(({ref}) => `<li>
      <span class="breadcrumb">${docs[ref].relUrl.split('/').filter(x => x && x !== 'docs').join(' > ')}</span>
      <a href="${docs[ref].url}">${docs[ref].title}</a>
    </li>`).join('') : '<li class="no-results">No Results</li>'
  }

  input.addEventListener('keyup', handler)
  input.addEventListener('input', handler)
  input.addEventListener('focus', () => searchContainer.classList.add('active'))
  input.addEventListener('blur', () => searchContainer.classList.remove('active'))
})();
