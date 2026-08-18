package admin

import (
	"context"
	"html/template"
	"net/url"
	"strconv"

	"chameth.com/chameth.com/features/admin/crud"
	"chameth.com/chameth.com/features/projects"
	"chameth.com/chameth.com/features/projects/admin/templates"
	"chameth.com/chameth.com/features/routing"
)

// projectWithSection carries the section name and ordering position that the
// list page shows alongside each project.
type projectWithSection struct {
	projects.Project
	SectionName  string
	SectionOrder int
}

func RegisterRoutes(rm *routing.Manager) {
	crud.Register(rm.Admin, "/projects", crud.Routes{
		List:   crud.List("project", fetchProjects, toSummary, templates.RenderListProjects),
		Create: crud.Create("project", "/projects", crud.GenerateName(projects.CreateProject)),
		Edit:   crud.Edit("project", projects.GetProjectByID, toEditData, templates.RenderEditProject),
		Update: crud.Update("project", "/projects", applyUpdate),
	})
}

func fetchProjects(ctx context.Context) ([]projectWithSection, []projectWithSection, error) {
	sections, err := projects.GetAllProjectSections(ctx)
	if err != nil {
		return nil, nil, err
	}

	names := make(map[int]string, len(sections))
	orders := make(map[int]int, len(sections))
	for i, section := range sections {
		names[section.ID] = section.Name
		orders[section.ID] = i
	}

	annotate := func(list []projects.Project) []projectWithSection {
		annotated := make([]projectWithSection, len(list))
		for i, project := range list {
			annotated[i] = projectWithSection{
				Project:      project,
				SectionName:  names[project.Section],
				SectionOrder: orders[project.Section],
			}
		}
		return annotated
	}

	draftProjects, err := projects.GetDraftProjects(ctx)
	if err != nil {
		return nil, nil, err
	}
	allProjects, err := projects.GetAllProjects(ctx)
	if err != nil {
		return nil, nil, err
	}

	return annotate(draftProjects), annotate(allProjects), nil
}

func toSummary(project projectWithSection) templates.ProjectSummary {
	return templates.ProjectSummary{
		ID:          project.ID,
		Name:        project.Name,
		Icon:        template.HTML(project.Icon),
		Pinned:      project.Pinned,
		Section:     project.SectionName,
		Description: project.Description,
		SectionSort: project.SectionOrder,
	}
}

func toEditData(ctx context.Context, project *projects.Project) (templates.EditProjectData, error) {
	sections, err := projects.GetAllProjectSections(ctx)
	if err != nil {
		return templates.EditProjectData{}, err
	}

	sectionOptions := make([]templates.SectionOption, len(sections))
	for i, section := range sections {
		sectionOptions[i] = templates.SectionOption{
			ID:   section.ID,
			Name: section.Name,
		}
	}

	return templates.EditProjectData{
		ID:                project.ID,
		Name:              project.Name,
		Icon:              project.Icon,
		Description:       project.Description,
		Section:           project.Section,
		Pinned:            project.Pinned,
		Published:         project.Published,
		AvailableSections: sectionOptions,
	}, nil
}

func applyUpdate(ctx context.Context, id int, form url.Values) error {
	section, err := strconv.Atoi(form.Get("section"))
	if err != nil {
		return err
	}

	return projects.UpdateProject(ctx, id,
		form.Get("name"),
		form.Get("icon"),
		form.Get("description"),
		section,
		form.Get("pinned") == "true",
		form.Get("published") == "true",
	)
}
