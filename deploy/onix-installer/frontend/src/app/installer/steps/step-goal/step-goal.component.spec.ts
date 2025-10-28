/**
 * Copyright 2025 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { ComponentFixture, TestBed, fakeAsync, tick } from '@angular/core/testing';
import { ReactiveFormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { RouterTestingModule } from '@angular/router/testing';
import { of, BehaviorSubject } from 'rxjs';
import { InstallerStateService } from '../../../core/services/installer-state.service';
import { StepGoalComponent } from './step-goal.component';
import { InstallerGoal, DeploymentGoal } from '../../types/installer.types';
import { ChangeDetectorRef } from '@angular/core';
import { HttpClientTestingModule } from '@angular/common/http/testing';

class MockInstallerStateService {
  private state = new BehaviorSubject<any>({
    installerGoal: null,
    deploymentGoal: null,
    deploymentStatus: null,
  });

  getCurrentState() {
    return this.state.getValue();
  }

  updateState(newState: any) {
    this.state.next({ ...this.state.getValue(), ...newState });
  }

  updateInstallerGoal(goal: InstallerGoal) {
    this.updateState({ installerGoal: goal });
  }

  updateDeploymentGoal(goal: DeploymentGoal) {
    this.updateState({ deploymentGoal: goal });
  }
}

describe('StepGoalComponent', () => {
  let component: StepGoalComponent;
  let fixture: ComponentFixture<StepGoalComponent>;
  let router: Router;
  let installerStateService: MockInstallerStateService;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [
        ReactiveFormsModule,
        RouterTestingModule.withRoutes([]),
        HttpClientTestingModule,
        StepGoalComponent,
      ],
      providers: [
        { provide: InstallerStateService, useClass: MockInstallerStateService },
        ChangeDetectorRef,
      ],
    }).compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(StepGoalComponent);
    component = fixture.componentInstance;
    router = TestBed.inject(Router);
    installerStateService = TestBed.inject(InstallerStateService) as unknown as MockInstallerStateService;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should initialize forms', () => {
    expect(component.goalForm).toBeDefined();
    expect(component.createNetworkComponentsForm).toBeDefined();
    expect(component.joinNetworkComponentsForm).toBeDefined();
  });

  it('should select "create_new_open_network" goal and show create form', () => {
    component.goalForm.get('installerGoal')?.setValue('create_new_open_network');
    fixture.detectChanges();
    expect(component.selectedInstallerGoal).toBe('create_new_open_network');
    const createForm = fixture.nativeElement.querySelector('#create-form');
    const joinForm = fixture.nativeElement.querySelector('#join-form');
    expect(createForm).toBeNull();
    expect(joinForm).toBeNull();
  });

  it('should select "join_existing_network" goal and show join form', () => {
    component.goalForm.get('installerGoal')?.setValue('join_existing_network');
    fixture.detectChanges();
    expect(component.selectedInstallerGoal).toBe('join_existing_network');
    const createForm = fixture.nativeElement.querySelector('#create-form');
    const joinForm = fixture.nativeElement.querySelector('#join-form');
    expect(createForm).toBeNull();
    expect(joinForm).toBeNull();
  });

  it('should enable next button when goal and components are selected for "create"', fakeAsync(() => {
    component.goalForm.get('installerGoal')?.setValue('create_new_open_network');
    component.createNetworkComponentsForm.get('registry')?.setValue(true);
    fixture.detectChanges();
    tick();
    expect(component.isGoalStepValid).toBeTrue();
  }));

  it('should enable next button when goal and components are selected for "join"', fakeAsync(() => {
    component.goalForm.get('installerGoal')?.setValue('join_existing_network');
    component.joinNetworkComponentsForm.get('bap')?.setValue(true);
    fixture.detectChanges();
    tick();
    expect(component.isGoalStepValid).toBeTrue();
  }));

  it('should disable next button when no goal is selected', fakeAsync(() => {
    fixture.detectChanges();
    tick();
    expect(component.isGoalStepValid).toBeFalse();
  }));

  it('should disable next button when goal is selected but no components for "create"', fakeAsync(() => {
    component.goalForm.get('installerGoal')?.setValue('create_new_open_network');
    fixture.detectChanges();
    tick();
    expect(component.isGoalStepValid).toBeFalse();
  }));

  it('should disable next button when goal is selected but no components for "join"', fakeAsync(() => {
    component.goalForm.get('installerGoal')?.setValue('join_existing_network');
    fixture.detectChanges();
    tick();
    expect(component.isGoalStepValid).toBeFalse();
  }));

  it('should navigate to prerequisites on next step', () => {
    const navigateSpy = spyOn(router, 'navigate');
    component.goalForm.get('installerGoal')?.setValue('create_new_open_network');
    component.createNetworkComponentsForm.get('gateway')?.setValue(true);
    fixture.detectChanges();
    component.goToNextStep();
    expect(navigateSpy).toHaveBeenCalledWith(['installer', 'prerequisites']);
  });

  it('should save state on next step', () => {
    const updateInstallerGoalSpy = spyOn(installerStateService, 'updateInstallerGoal').and.callThrough();
    const updateDeploymentGoalSpy = spyOn(installerStateService, 'updateDeploymentGoal').and.callThrough();
    component.goalForm.get('installerGoal')?.setValue('create_new_open_network');
    component.createNetworkComponentsForm.get('bap')?.setValue(true);
    fixture.detectChanges();
    component.goToNextStep();
    expect(updateInstallerGoalSpy).toHaveBeenCalledWith('create_new_open_network');
    expect(updateDeploymentGoalSpy).toHaveBeenCalledWith({
      registry: false,
      gateway: false,
      bap: true,
      bpp: false,
    });
  });

  it('should load initial state', () => {
    installerStateService.updateState({
      installerGoal: 'join_existing_network',
      deploymentGoal: { bap: true, bpp: false },
    });
    component.ngOnInit();
    fixture.detectChanges();
    expect(component.goalForm.get('installerGoal')?.value).toBe('join_existing_network');
    expect(component.joinNetworkComponentsForm.get('bap')?.value).toBe(true);
    expect(component.isGoalStepValid).toBeTrue();
  });

  it('should toggle all checkboxes for "create" form', () => {
    component.goalForm.get('installerGoal')?.setValue('create_new_open_network');
    fixture.detectChanges();

    const event = { checked: true };
    component.toggleSelectAll(event, 'create');
    fixture.detectChanges();

    expect(component.createNetworkComponentsForm.get('registry')?.value).toBe(false);
    expect(component.createNetworkComponentsForm.get('gateway')?.value).toBe(false);
    expect(component.createNetworkComponentsForm.get('bap')?.value).toBe(false);
    expect(component.createNetworkComponentsForm.get('bpp')?.value).toBe(false);
  });

  it('should toggle all checkboxes for "join" form', () => {
    component.goalForm.get('installerGoal')?.setValue('join_existing_network');
    fixture.detectChanges();

    const event = { checked: true };
    component.toggleSelectAll(event, 'join');
    fixture.detectChanges();

    expect(component.joinNetworkComponentsForm.get('bap')?.value).toBe(false);
    expect(component.joinNetworkComponentsForm.get('bpp')?.value).toBe(false);
  });
});
