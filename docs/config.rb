###
# Page options, layouts, aliases and proxies
###

# Per-page layout changes:
#
# With no layout
page '/*.xml', layout: false
page '/*.json', layout: false
page '/*.txt', layout: false

# With alternative layout
# page "/path/to/file.html", layout: :otherlayout

# Proxy pages (http://middlemanapp.com/basics/dynamic-pages/)
# proxy "/this-page-has-no-template.html", "/template-file.html", locals: {
#  which_fake_page: "Rendering a fake page with a local variable" }

# General configuration
activate :asciidoc

set :asciidoc_attributes, %w(icons=font)

# Reload the browser automatically whenever files change
configure :development do
  activate :livereload
end

class CustomMarkdown < Middleman::Extension
  $markdown_options = {
    autolink:           true,
    fenced_code_blocks: true,
    no_intra_emphasis:  true,
    strikethrough:      true,
    tables:             true,
    hard_wrap:          false,
    with_toc_data:      true
  }

  # Markdown files
  def initialize(app, options_hash={}, &block)
    super
    app.set :markdown_engine, :redcarpet
    app.set :markdown, $markdown_options
  end

  # TOC helper
  helpers do
    # Based on https://github.com/vmg/redcarpet/pull/186#issuecomment-22783188
    def toc(page)
      html_toc = Redcarpet::Markdown.new(Redcarpet::Render::HTML_TOC)
      file = ::File.read(page.source_file)

      # remove YAML frontmatter
      file = file.gsub(/^(---\s*\n.*?\n?)^(---\s*$\n?)/m,'')

      # quick fix for HAML: remove :markdown filter and indentation
      file = file.gsub(/:markdown\n/,'')
      file = file.gsub(/\t/,'')

      html_toc.render file
    end
  end
end

::Middleman::Extensions.register(:custom_markdown, CustomMarkdown)

activate :custom_markdown
activate :syntax, :line_numbers => true

# Build-specific configuration
configure :build do
  # Minify CSS on build
  # activate :minify_css

  # Minify Javascript on build
  # activate :minify_javascript
end
