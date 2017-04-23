require 'yaml'

module CustomHelpers

  def navigation_menu
    navigation = ""
    topic_map = YAML.load_file('build/_topic_map.yml')
    topic_map["Topics"].each_with_index do |topic, index|
      topic_directory = topic["Dir"]
      if topic_directory.nil?
        next "<li><a class=\"\" href=\"/#{topic["File"]}.html\">&nbsp;#{topic["Name"]}</a></li>"
      end

      caret = "fa-caret-right"
      in_class = ""
      subnav = ""
      entries = ""
      sub_topics = topic["Topics"]
      unless sub_topics.nil?
        sub_topics.each do |subtopic|
          path = current_page.url
          sub_topic_file = subtopic["File"]
          sub_topic_name = subtopic["Name"]
          if !path.end_with?("html")
            path += "index.html"
          end
          if path == "/#{topic_directory}/#{sub_topic_file}.html"
            in_class = "in"
            caret = "fa-caret-down"
            entries += "<li><a class=\"active\" href=\"/#{topic_directory}/#{sub_topic_file}.html\">&nbsp;#{sub_topic_name}</a></li>"
          else
            entries += "<li><a class=\"\" href=\"/#{topic_directory}/#{sub_topic_file}.html\">&nbsp;#{sub_topic_name}</a></li>"
          end
        end

        subnav += "<ul id=\"topicSubGroup-#{index}\" class=\"nav-tertiary list-unstyled collapse #{in_class}\">"
        subnav += entries
        subnav += "</ul>"
      end

      topic_nav = <<-HEREDOC
      <li class="nav-header">
        <a class="" href="javascript:void(0);" data-toggle="collapse" data-target="#topicSubGroup-%d">
          <span id="sgSpan-%d" class="fa %s"></span>&nbsp;%s
        </a>
      %s
      </li>
      HEREDOC
      navigation += topic_nav % [index, index, caret, topic["Name"], subnav]
    end

    navigation
  end
end
